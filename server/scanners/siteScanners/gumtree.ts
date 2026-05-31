import type { Page } from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import { db } from "../../globals/PrismaClient.js";
import { resultAsync } from "../../functions/neverthrowUtils.js";
import type { DealNotification } from "../../deals/types.js";
import {
  evaluateListingForQuery,
  matchPriceRange,
  persistListingObservation,
  type DiscoveredListing,
} from "../listingState.js";

interface GumtreeQueryInput {
  url: string;
}

interface GumtreeValues {
  readonly foundName: string;
  readonly foundPrice: number;
  readonly foundImg: string | null;
  readonly foundUrl: string;
}

const gumtreeCardSelector =
  "a[class='user-ad-row-new-design link link--base-color-inherit link--hover-color-none link--no-underline'], a[href*='/s-ad/']";

export async function scanGumtree(page: Page): Promise<DealNotification[]> {
  if (!globals.GUMTREE) return [];
  setStatus("Scanning Gumtree");

  const queries = await db.gumtree.findMany();
  const notifications: DealNotification[] = [];

  for (const item of queries) {
    const queryId = await ensureGumtreeQueryId(item.url, item.queryId);
    const foundItems = await getGumtreeListings(page, item);

    for (const foundItem of foundItems) {
      const listing = normalizeGumtreeListing(foundItem);
      const persistedResult = await persistListingObservation({ listing });
      if (persistedResult.isErr()) {
        console.warn(persistedResult.error.message);
        continue;
      }

      const decisionResult = await evaluateListingForQuery({
        queryId,
        listingId: persistedResult.value.listingId,
        source: "gumtree",
        totalPrice: listing.totalPrice,
        match: matchPriceRange(listing.totalPrice, null, item.maxPrice),
      });
      if (decisionResult.isErr()) {
        console.warn(decisionResult.error.message);
        continue;
      }

      if (decisionResult.value.type === "Notify") {
        notifications.push({
          kind: "deal",
          source: "gumtree",
          title: `a ${listing.title} priced at $${listing.totalPrice ?? "unknown"} is available at ${listing.canonicalUrl}`,
          url: listing.canonicalUrl,
          price: listing.totalPrice ?? undefined,
          ...(listing.imageUrl && { imageUrl: listing.imageUrl }),
          query: {
            type: "gumtree",
            id: item.url,
          },
        });
      }
    }
  }

  return notifications;
}

export async function getGumtreeListings(
  page: Page,
  item: GumtreeQueryInput,
): Promise<GumtreeValues[]> {
  await page.setUserAgent(
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
  );
  const gotoResult = await resultAsync(
    () => page.goto(item.url, { waitUntil: "domcontentloaded", timeout: 10000 }),
    "Gumtree navigation failed",
  );
  if (gotoResult.isErr()) {
    console.warn(gotoResult.error.message);
    return [];
  }

  const selectorResult = await resultAsync(
    () => page.waitForSelector(gumtreeCardSelector, { timeout: 10000 }),
    "Gumtree selector lookup failed",
  );
  if (selectorResult.isErr()) {
    console.warn(selectorResult.error.message);
    return [];
  }

  const extractionResult = await resultAsync(
    () =>
      page.$$eval(gumtreeCardSelector, (cards): GumtreeValues[] =>
        cards
          .map((card): GumtreeValues | null => {
            const parsePrice = (input: string | null): number | null => {
              if (!input) return null;
              if (input.toLocaleLowerCase().includes("free")) return 0;
              const value = Number.parseFloat(input.replace(/[^0-9.]+/g, ""));
              return Number.isFinite(value) ? value : null;
            };
            const link =
              card instanceof HTMLAnchorElement ? card.href : null;
            const title =
              card
                .querySelector(
                  "div.user-ad-row-new-design__main-content > p.user-ad-row-new-design__title > span, [class*='title']",
                )
                ?.textContent?.trim() ?? null;
            const priceText =
              card
                .querySelector(
                  "div.user-ad-row-new-design__right-content span.user-ad-price-new-design__price, [class*='price']",
                )
                ?.textContent?.trim() ?? null;
            const imageElement = card.querySelector("img");
            const foundImg =
              imageElement instanceof HTMLImageElement ? imageElement.src : null;
            const foundPrice = parsePrice(priceText);

            if (!link || !title || foundPrice === null) return null;
            return {
              foundName: title,
              foundPrice,
              foundImg,
              foundUrl: link,
            };
          })
          .filter((card): card is GumtreeValues => card !== null),
      ),
    "Gumtree listing extraction failed",
  );
  if (extractionResult.isErr()) {
    console.warn(extractionResult.error.message);
    return [];
  }

  return dedupeByUrl(extractionResult.value);
}

export async function getGumtreeValues(page: Page, item: GumtreeQueryInput) {
  const listings = await getGumtreeListings(page, item);
  return listings[0] ?? null;
}

function normalizeGumtreeListing(item: GumtreeValues): DiscoveredListing {
  return {
    source: "gumtree",
    externalId: null,
    canonicalUrl: item.foundUrl,
    title: item.foundName,
    price: item.foundPrice,
    shipping: null,
    totalPrice: item.foundPrice,
    currency: "AUD",
    imageUrl: item.foundImg,
    description: null,
    availability: "available",
  };
}

async function ensureGumtreeQueryId(
  url: string,
  queryId: string | null,
): Promise<string> {
  if (queryId) return queryId;
  const query = await db.query.create({
    data: {
      gumtree: {
        connect: { url },
      },
    },
  });
  return query.id;
}

function dedupeByUrl(items: GumtreeValues[]): GumtreeValues[] {
  const seen = new Set<string>();
  const results: GumtreeValues[] = [];
  for (const item of items) {
    if (seen.has(item.foundUrl)) continue;
    seen.add(item.foundUrl);
    results.push(item);
  }
  return results;
}
