import { Page } from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import selectorRace from "../../functions/selectorRace.js";
import { db } from "../../globals/PrismaClient.js";
import { resultAsync } from "../../functions/neverthrowUtils.js";
import type { DealNotification } from "../../deals/types.js";
import {
  evaluateListingForQuery,
  matchPriceRange,
  persistListingObservation,
  type DiscoveredListing,
} from "../listingState.js";

interface EbayQueryInput {
  url: string;
}

export interface RawEbayItem {
  title: string;
  priceText: string | null;
  imageUrl: string | null;
  url: string | null;
}

interface EbayCardSelectorResult {
  readonly textContent?: string | null;
  readonly src?: string | null;
  readonly href?: string | null;
}

export interface EbayCardElement {
  querySelector(selectors: string): EbayCardSelectorResult | null;
}

interface EbayValues {
  foundName: string | null;
  foundPrice: number | null;
  foundImage: string | null;
  foundUrl: string | null;
  foundCurrency: string | null;
  externalId: string | null;
}

export const ebaySelectors: {
  readonly resultsContainer: string;
  readonly emptyResults: string;
  readonly resultCard: string;
  readonly title: string;
  readonly price: string;
  readonly image: string;
  readonly link: string;
} = {
  resultsContainer: "div.srp-river",
  emptyResults: ".srp-save-null-search__heading",
  resultCard:
    "div.su-card-container.su-card-container--horizontal, li.s-item",
  title: 'div[role="heading"], .s-item__title',
  price: "span.s-card__price, .s-item__price",
  image: "img",
  link: "a[href]",
};

function parseEbayPrice(priceText: string | null): {
  readonly price: number | null;
  readonly currency: string | null;
} {
  if (!priceText) return { price: null, currency: null };
  if (priceText.startsWith("AU $")) {
    return {
      price: parseFloat(priceText.replace(/[^0-9.-]+/g, "")),
      currency: "AUD",
    };
  }

  // Hack currency conversion
  return {
    price: parseFloat(priceText.replace(/[^0-9.-]+/g, "")) * 1.55,
    currency: "USD",
  };
}

function parseEbayName(nameText: string | null): string | null {
  if (!nameText) return null;
  if (nameText.startsWith("New listing")) {
    return nameText.replace("New listing", "").trim();
  }
  return nameText.trim();
}

export function parseEbayItem(rawItem: RawEbayItem | null): EbayValues {
  if (!rawItem) {
    return emptyEbayValues();
  }

  const foundPrice = parseEbayPrice(rawItem.priceText);
  const parsedName = parseEbayName(rawItem.title);
  const externalId = parseEbayExternalId(rawItem.url);

  if (!parsedName || foundPrice.price === null) {
    return emptyEbayValues();
  }

  return {
    foundName: parsedName,
    foundPrice: foundPrice.price,
    foundImage: rawItem.imageUrl,
    foundUrl: rawItem.url,
    foundCurrency: foundPrice.currency,
    externalId,
  };
}

export function readRawEbayItem(card: EbayCardElement): RawEbayItem {
  const title =
    card.querySelector(ebaySelectors.title)?.textContent?.trim() ?? "";
  const priceText =
    card.querySelector(ebaySelectors.price)?.textContent?.trim() ?? null;
  const imageUrl = card.querySelector(ebaySelectors.image)?.src ?? null;
  const url = card.querySelector(ebaySelectors.link)?.href ?? null;

  return { title, priceText, imageUrl, url };
}

export async function scanEbay(page: Page): Promise<DealNotification[]> {
  if (!globals.EBAY) return [];

  setStatus("Scanning eBay");

  const queries = await db.ebay.findMany();
  const notifications: DealNotification[] = [];

  for (const item of queries) {
    const queryId = await ensureEbayQueryId(item.url, item.queryId);
    const foundItems = await getEbayListings(page, item);

    for (const foundItem of foundItems) {
      const listing = normalizeEbayListing(foundItem);
      const persistedResult = await persistListingObservation({ listing });
      if (persistedResult.isErr()) {
        console.warn(persistedResult.error.message);
        continue;
      }

      const decisionResult = await evaluateListingForQuery({
        queryId,
        listingId: persistedResult.value.listingId,
        source: "ebay",
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
          source: "ebay",
          title: `a ${listing.title} priced at $${listing.totalPrice ?? "unknown"} is available at ${listing.canonicalUrl}`,
          url: listing.canonicalUrl,
          price: listing.totalPrice ?? undefined,
          imageUrl: listing.imageUrl ?? undefined,
          query: {
            type: "ebay",
            id: item.url,
          },
        });
      }
    }
  }
  return notifications;
}

export async function getEbayValues(page: Page, item: EbayQueryInput) {
  const gotoResult = await resultAsync(
    () =>
      page.goto(item.url, { waitUntil: "domcontentloaded", timeout: 10000 }),
    "eBay navigation failed",
  );
  if (gotoResult.isErr()) {
    console.warn(gotoResult.error.message);
    return emptyEbayValues();
  }

  const selector = await selectorRace(
    page,
    ebaySelectors.resultsContainer,
    ebaySelectors.emptyResults,
  );
  if (!selector) return emptyEbayValues();

  const rawItemsResult = await resultAsync(
    () =>
      selector.$$eval(
        ebaySelectors.resultCard,
        (cards, selectors): RawEbayItem[] =>
          cards.map((card) => {
            const title =
              card.querySelector(selectors.title)?.textContent?.trim() ?? "";
            const priceText =
              card.querySelector(selectors.price)?.textContent?.trim() ?? null;
            const imageUrl =
              card.querySelector<HTMLImageElement>(selectors.image)?.src ??
              null;
            const url =
              card.querySelector<HTMLAnchorElement>(selectors.link)?.href ??
              null;

            return { title, priceText, imageUrl, url };
          }),
        ebaySelectors,
      ),
    "eBay result card extraction failed",
  );
  if (rawItemsResult.isErr()) {
    console.warn(rawItemsResult.error.message);
    return emptyEbayValues();
  }

  return parseEbayItem(
    rawItemsResult.value.find((rawItem) => rawItem.title) ?? null,
  );
}

export async function getEbayListings(
  page: Page,
  item: EbayQueryInput,
): Promise<EbayValues[]> {
  const gotoResult = await resultAsync(
    () =>
      page.goto(item.url, { waitUntil: "domcontentloaded", timeout: 10000 }),
    "eBay navigation failed",
  );
  if (gotoResult.isErr()) {
    console.warn(gotoResult.error.message);
    return [];
  }

  const selector = await selectorRace(
    page,
    ebaySelectors.resultsContainer,
    ebaySelectors.emptyResults,
  );
  if (!selector) return [];

  const rawItemsResult = await resultAsync(
    () =>
      selector.$$eval(
        ebaySelectors.resultCard,
        (cards, selectors): RawEbayItem[] =>
          cards.map((card) => {
            const title =
              card.querySelector(selectors.title)?.textContent?.trim() ?? "";
            const priceText =
              card.querySelector(selectors.price)?.textContent?.trim() ?? null;
            const imageElement = card.querySelector(selectors.image);
            const imageUrl =
              imageElement instanceof HTMLImageElement ? imageElement.src : null;
            const linkElement = card.querySelector(selectors.link);
            const url =
              linkElement instanceof HTMLAnchorElement ? linkElement.href : null;

            return { title, priceText, imageUrl, url };
          }),
        ebaySelectors,
      ),
    "eBay result card extraction failed",
  );
  if (rawItemsResult.isErr()) {
    console.warn(rawItemsResult.error.message);
    return [];
  }

  return rawItemsResult.value
    .filter((rawItem) => isEbayListingUrl(rawItem.url))
    .map((rawItem) => parseEbayItem(rawItem))
    .filter(
      (item) =>
        item.foundName !== null &&
        item.foundPrice !== null &&
        item.foundUrl !== null,
    );
}

function normalizeEbayListing(item: EbayValues): DiscoveredListing {
  return {
    source: "ebay",
    externalId: item.externalId,
    canonicalUrl: item.foundUrl ?? "",
    title: item.foundName ?? "",
    price: item.foundPrice,
    shipping: null,
    totalPrice: item.foundPrice,
    currency: item.foundCurrency,
    imageUrl: item.foundImage,
    description: null,
    availability: "available",
  };
}

async function ensureEbayQueryId(url: string, queryId: string | null): Promise<string> {
  if (queryId) return queryId;
  const query = await db.query.create({
    data: {
      ebay: {
        connect: { url },
      },
    },
  });
  return query.id;
}

function emptyEbayValues(): EbayValues {
  return {
    foundName: null,
    foundPrice: null,
    foundImage: null,
    foundUrl: null,
    foundCurrency: null,
    externalId: null,
  };
}

function isEbayListingUrl(url: string | null): boolean {
  return url !== null && /\/itm\//.test(url);
}

function parseEbayExternalId(url: string | null): string | null {
  if (!url) return null;
  const match = /\/itm\/(?:[^/?#]+\/)?(\d+)/.exec(url);
  return match?.[1] ?? null;
}
