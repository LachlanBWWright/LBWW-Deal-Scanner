import type { Page } from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import { formatPrice } from "../../functions/formatPrice.js";
import { db } from "../../globals/PrismaClient.js";
import { resultAsync, resultSync } from "../../functions/neverthrowUtils.js";
import type { DealNotification } from "../../deals/types.js";
import {
  combineMatchResults,
  evaluateListingForQuery,
  matchPhrases,
  matchPriceRange,
  persistListingObservation,
  type DiscoveredListing,
} from "../listingState.js";

interface CashConvertersQueryInput {
  url: string;
  requiredPhrases?: string;
  excludePhrases?: string;
  maxPrice?: number | null;
  queryId?: string | null;
}

export interface CashConvertersSummary {
  readonly canonicalUrl: string;
  readonly title: string;
  readonly price: number | null;
  readonly shipping: number | null;
  readonly totalPrice: number | null;
  readonly imageUrl: string | null;
}

interface CashConvertersDetail {
  readonly canonicalUrl: string;
  readonly title: string;
  readonly description: string | null;
  readonly price: number | null;
  readonly shipping: number | null;
  readonly totalPrice: number | null;
  readonly imageUrl: string | null;
  readonly availability: string | null;
}

export async function scanCashConverters(
  page: Page,
): Promise<DealNotification[]> {
  if (!globals.CASH_CONVERTERS) return [];
  setStatus("Scanning Cash Converters");

  const queries = await db.cashConverters.findMany();
  const notifications: DealNotification[] = [];

  for (const query of queries) {
    if (query.scanMode === "siteWide") continue;
    const queryId = await ensureCashConvertersQueryId(query.url, query.queryId);
    const summaries = await discoverCashConvertersItems(page, query);
    notifications.push(
      ...(await reviewCashConvertersSummaries(page, summaries, [query], queryId)),
    );
  }

  if (isCashConvertersSiteWideEnabled()) {
    const siteWideQueries = queries.filter((query) => query.scanMode === "siteWide");
    const discoveryUrls = getCashConvertersSiteWideUrls(siteWideQueries);
    for (const discoveryUrl of discoveryUrls) {
      const summaries = await discoverCashConvertersItems(page, { url: discoveryUrl });
      for (const query of siteWideQueries) {
        const queryId = await ensureCashConvertersQueryId(query.url, query.queryId);
        notifications.push(
          ...(await reviewCashConvertersSummaries(
            page,
            summaries,
            [query],
            queryId,
          )),
        );
      }
    }
  }

  return notifications;
}

async function reviewCashConvertersSummaries(
  page: Page,
  summaries: CashConvertersSummary[],
  queries: CashConvertersQueryInput[],
  queryId: string,
): Promise<DealNotification[]> {
  const notifications: DealNotification[] = [];
  const concurrency = getPositiveIntegerEnv(
    "CASH_CONVERTERS_DETAIL_CONCURRENCY",
    1,
  );

  for (let index = 0; index < summaries.length; index += concurrency) {
    const batch = summaries.slice(index, index + concurrency);
    const batchNotifications = await Promise.all(
      batch.map((summary) =>
        reviewCashConvertersSummary(page, summary, queries, queryId),
      ),
    );
    notifications.push(...batchNotifications.flat());
  }

  return notifications;
}

async function reviewCashConvertersSummary(
  rootPage: Page,
  summary: CashConvertersSummary,
  queries: CashConvertersQueryInput[],
  queryId: string,
): Promise<DealNotification[]> {
  const pageResult = await resultAsync(
    () => rootPage.browser().newPage(),
    "Failed to create Cash Converters detail page",
  );
  if (pageResult.isErr()) {
    console.warn(pageResult.error.message);
    return [];
  }

  const page = pageResult.value;
  const detail = await getCashConvertersDetailForSummary(page, summary);
  const closeResult = await resultAsync(
    () => page.close(),
    "Failed to close Cash Converters detail page",
  );
  if (closeResult.isErr()) {
    console.warn(closeResult.error.message);
  }

  const listing = normalizeCashConvertersListing(detail);
  const persistedResult = await persistListingObservation({
    listing,
    detailFetchedAt: detail.description === null ? null : new Date(),
  });
  if (persistedResult.isErr()) {
    console.warn(persistedResult.error.message);
    return [];
  }

  const notifications: DealNotification[] = [];
  for (const query of queries) {
    const searchableText = `${listing.title}\n${listing.description ?? ""}`;
    const match = combineMatchResults(
      matchPhrases(
        searchableText,
        query.requiredPhrases,
        query.excludePhrases,
      ),
      matchPriceRange(listing.totalPrice, null, query.maxPrice),
    );
    const decisionResult = await evaluateListingForQuery({
      queryId,
      listingId: persistedResult.value.listingId,
      source: "cashConverters",
      totalPrice: listing.totalPrice,
      match,
    });
    if (decisionResult.isErr()) {
      console.warn(decisionResult.error.message);
      continue;
    }

    if (decisionResult.value.type === "Notify") {
      notifications.push(
        buildCashConvertersNotification(query, listing, decisionResult.value.reason),
      );
    }
  }

  return notifications;
}

export async function discoverCashConvertersItems(
  page: Page,
  item: CashConvertersQueryInput,
): Promise<CashConvertersSummary[]> {
  const pageLimit = getPositiveIntegerEnv(
    "CASH_CONVERTERS_SITEWIDE_PAGE_LIMIT",
    1,
  );
  const results: CashConvertersSummary[] = [];

  for (let pageNumber = 1; pageNumber <= pageLimit; pageNumber++) {
    const pageUrlResult = withCashConvertersPage(item.url, pageNumber);
    if (pageUrlResult.isErr()) {
      console.warn(pageUrlResult.error.message);
      break;
    }

    const pageResults = await discoverCashConvertersPageItems(
      page,
      pageUrlResult.value,
    );
    if (pageResults.length === 0) break;
    results.push(...pageResults);
  }

  return dedupeByUrl(results);
}

async function discoverCashConvertersPageItems(
  page: Page,
  url: string,
): Promise<CashConvertersSummary[]> {
  const gotoResult = await resultAsync(
    () => page.goto(url, { waitUntil: "domcontentloaded", timeout: 15000 }),
    "Cash Converters navigation failed",
  );
  if (gotoResult.isErr()) {
    console.warn(gotoResult.error.message);
    return [];
  }

  const extractionResult = await resultAsync(
    () =>
      page.$$eval("div[class='product-item'], .product-item", (cards) =>
        cards
          .map((card): CashConvertersSummary | null => {
            const title =
              card
                .querySelector("span[class='product-item__title__description']")
                ?.textContent?.trim() ?? null;
            const priceText =
              card.querySelector(".product-item__price")?.textContent ?? null;
            const shippingText =
              card.querySelector(".product-item__postage")?.textContent ?? null;
            const imageElement = card.querySelector("img");
            const imageUrl =
              imageElement instanceof HTMLImageElement ? imageElement.src : null;
            const linkElement = card.querySelector("a");
            const canonicalUrl =
              linkElement instanceof HTMLAnchorElement ? linkElement.href : null;
            const price = parseCashConvertersPriceText(priceText);
            const shipping = parseCashConvertersShippingText(shippingText);
            const totalPrice =
              price === null ? null : price + (shipping === null ? 0 : shipping);

            if (!title || !canonicalUrl) return null;
            return {
              canonicalUrl,
              title,
              price,
              shipping,
              totalPrice,
              imageUrl,
            };
          })
          .filter((card): card is CashConvertersSummary => card !== null),
      ),
    "Cash Converters result card extraction failed",
  );
  if (extractionResult.isErr()) {
    console.warn(extractionResult.error.message);
    return [];
  }

  return extractionResult.value;
}

async function getCashConvertersDetailForSummary(
  page: Page,
  summary: CashConvertersSummary,
): Promise<CashConvertersDetail> {
  const existing = await db.listing.findUnique({
    where: {
      source_canonicalUrl: {
        source: "cashConverters",
        canonicalUrl: summary.canonicalUrl,
      },
    },
  });

  if (existing?.lastDetailAt && !isStaleDetail(existing.lastDetailAt)) {
    return {
      canonicalUrl: summary.canonicalUrl,
      title: existing.title,
      description: existing.description,
      price: summary.price,
      shipping: summary.shipping,
      totalPrice: summary.totalPrice,
      imageUrl: summary.imageUrl ?? existing.imageUrl,
      availability: existing.availability,
    };
  }

  return await fetchCashConvertersItemDetail(page, summary);
}

export async function fetchCashConvertersItemDetail(
  page: Page,
  summary: CashConvertersSummary,
): Promise<CashConvertersDetail> {
  const gotoResult = await resultAsync(
    () =>
      page.goto(summary.canonicalUrl, {
        waitUntil: "domcontentloaded",
        timeout: 15000,
      }),
    "Cash Converters detail navigation failed",
  );
  if (gotoResult.isErr()) {
    console.warn(gotoResult.error.message);
    return {
      ...summary,
      description: null,
      availability: null,
    };
  }

  const detailResult = await resultAsync(
    () =>
      page.evaluate((fallback): CashConvertersDetail => {
        const parsePrice = (input: string | null): number | null => {
          if (!input) return null;
          const value = Number.parseFloat(input.replace(/[^0-9.]+/g, ""));
          return Number.isFinite(value) ? value : null;
        };
        const title =
          document
            .querySelector(
              "h1, .product-detail__title, .product-title, [class*='product'][class*='title']",
            )
            ?.textContent?.trim() || fallback.title;
        const description =
          document
            .querySelector(
              ".product-detail__description, .product-description, [class*='description']",
            )
            ?.textContent?.replace(/\s+/g, " ")
            .trim() || null;
        const priceText =
          document
            .querySelector(".product-item__price, [class*='price']")
            ?.textContent ?? null;
        const imageElement = document.querySelector("img");
        const imageUrl =
          imageElement instanceof HTMLImageElement
            ? imageElement.src
            : fallback.imageUrl;
        const price = parsePrice(priceText) ?? fallback.price;
        const shipping = fallback.shipping;
        const totalPrice =
          price === null ? fallback.totalPrice : price + (shipping ?? 0);
        const unavailable = document.body.textContent
          ?.toLocaleLowerCase()
          .includes("sold");

        return {
          canonicalUrl: fallback.canonicalUrl,
          title,
          description,
          price,
          shipping,
          totalPrice,
          imageUrl,
          availability: unavailable ? "unavailable" : "available",
        };
      }, summary),
    "Cash Converters detail extraction failed",
  );
  if (detailResult.isErr()) {
    console.warn(detailResult.error.message);
    return {
      ...summary,
      description: null,
      availability: null,
    };
  }

  return detailResult.value;
}

export function normalizeCashConvertersListing(
  detail: CashConvertersDetail,
): DiscoveredListing {
  return {
    source: "cashConverters",
    externalId: null,
    canonicalUrl: detail.canonicalUrl,
    title: detail.title,
    price: detail.price,
    shipping: detail.shipping,
    totalPrice: detail.totalPrice,
    currency: "AUD",
    imageUrl: detail.imageUrl,
    description: detail.description,
    availability: detail.availability,
  };
}

export async function getCashConvertersValues(
  page: Page,
  item: CashConvertersQueryInput,
) {
  const summaries = await discoverCashConvertersItems(page, item);
  const first = summaries[0];
  if (!first) return;
  return {
    itemName: first.title,
    totalPrice: first.totalPrice ?? 0,
    image: first.imageUrl ?? "",
  };
}

function buildCashConvertersNotification(
  query: CashConvertersQueryInput,
  listing: DiscoveredListing,
  reason: "FirstMatch" | "PriceDroppedIntoRange" | "FurtherPriceDrop",
): DealNotification {
  const priceText =
    listing.totalPrice === null ? "an unknown price" : formatPrice(listing.totalPrice);
  const reasonText =
    reason === "PriceDroppedIntoRange"
      ? " after a price reduction"
      : reason === "FurtherPriceDrop"
        ? " after a further price reduction"
        : "";

  return {
    kind: "deal",
    source: "cashConverters",
    title: `a ${listing.title} for ${priceText} is available${reasonText} at ${listing.canonicalUrl}`,
    url: listing.canonicalUrl,
    price: listing.totalPrice ?? undefined,
    imageUrl: listing.imageUrl ?? undefined,
    query: {
      type: "cashConverters",
      id: query.url,
    },
  };
}

async function ensureCashConvertersQueryId(
  url: string,
  queryId: string | null,
): Promise<string> {
  if (queryId) return queryId;
  const query = await db.query.create({
    data: {
      cashConverters: {
        connect: { url },
      },
    },
  });
  return query.id;
}

function dedupeByUrl(items: CashConvertersSummary[]): CashConvertersSummary[] {
  const seen = new Set<string>();
  const results: CashConvertersSummary[] = [];
  for (const item of items) {
    if (seen.has(item.canonicalUrl)) continue;
    seen.add(item.canonicalUrl);
    results.push(item);
  }
  return results;
}

function parseCashConvertersPriceText(input: string | null): number | null {
  if (!input) return null;
  const value = Number.parseFloat(input.replace(/[^0-9.]+/g, ""));
  return Number.isFinite(value) ? value : null;
}

function parseCashConvertersShippingText(input: string | null): number | null {
  if (!input) return null;
  if (input.toLocaleLowerCase().includes("free")) return 0;
  return parseCashConvertersPriceText(input);
}

function isCashConvertersSiteWideEnabled() {
  return process.env.CASH_CONVERTERS_SITEWIDE_ENABLED === "true";
}

function getCashConvertersSiteWideUrls(
  queries: CashConvertersQueryInput[],
): string[] {
  const configuredUrls = String(
    process.env.CASH_CONVERTERS_SITEWIDE_URLS ?? "",
  )
    .split(",")
    .map((url) => url.trim())
    .filter((url) => url.length > 0);

  const urls =
    configuredUrls.length > 0 ? configuredUrls : queries.map((query) => query.url);
  return [...new Set(urls)];
}

function withCashConvertersPage(url: string, pageNumber: number) {
  return resultSync(() => {
    const parsed = new URL(url);
    parsed.searchParams.set("page", String(pageNumber));
    return parsed.toString();
  }, "Invalid Cash Converters URL");
}

function isStaleDetail(lastDetailAt: Date) {
  const staleHours = getPositiveIntegerEnv(
    "CASH_CONVERTERS_DETAIL_STALE_HOURS",
    24,
  );
  return Date.now() - lastDetailAt.getTime() > staleHours * 60 * 60 * 1000;
}

function getPositiveIntegerEnv(name: string, fallback: number) {
  const value = Number(process.env[name] ?? fallback);
  return Number.isInteger(value) && value > 0 ? value : fallback;
}
