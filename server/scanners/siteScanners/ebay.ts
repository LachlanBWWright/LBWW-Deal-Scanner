import { Page } from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import selectorRace from "../../functions/selectorRace.js";
import { db, SCANNER } from "../../globals/PrismaClient.js";
import { checkIfNew } from "../../functions/handleItemUpdate.js";
import { resultAsync } from "../../functions/neverthrowUtils.js";
import type { DealNotification } from "../../deals/types.js";

interface EbayQueryInput {
  url: string;
}

export interface RawEbayItem {
  title: string;
  priceText: string | null;
  imageUrl: string | null;
}

interface EbayCardSelectorResult {
  readonly textContent?: string | null;
  readonly src?: string | null;
}

export interface EbayCardElement {
  querySelector(selectors: string): EbayCardSelectorResult | null;
}

interface EbayValues {
  foundName: string | null;
  foundPrice: number | null;
  foundImage: string | null;
}

export const ebaySelectors: {
  readonly resultsContainer: string;
  readonly emptyResults: string;
  readonly resultCard: string;
  readonly title: string;
  readonly price: string;
  readonly image: string;
} = {
  resultsContainer: "div.srp-river",
  emptyResults: ".srp-save-null-search__heading",
  resultCard:
    "div.su-card-container.su-card-container--horizontal, li.s-item",
  title: 'div[role="heading"], .s-item__title',
  price: "span.s-card__price, .s-item__price",
  image: "img",
};

let isAud = true;

function parseEbayPrice(priceText: string | null): number | null {
  if (!priceText) return null;

  if (priceText.startsWith("AU $")) {
    isAud = true;
    return parseFloat(priceText.replace(/[^0-9.-]+/g, ""));
  }

  isAud = false;
  // Hack currency conversion
  return parseFloat(priceText.replace(/[^0-9.-]+/g, "")) * 1.55;
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
    return { foundName: null, foundPrice: null, foundImage: null };
  }

  const foundPrice = parseEbayPrice(rawItem.priceText);
  const parsedName = parseEbayName(rawItem.title);

  if (!parsedName || !foundPrice || !rawItem.imageUrl) {
    return { foundName: null, foundPrice: null, foundImage: null };
  }

  return {
    foundName: parsedName,
    foundPrice,
    foundImage: rawItem.imageUrl,
  };
}

export function readRawEbayItem(card: EbayCardElement): RawEbayItem {
  const title =
    card.querySelector(ebaySelectors.title)?.textContent?.trim() ?? "";
  const priceText =
    card.querySelector(ebaySelectors.price)?.textContent?.trim() ?? null;
  const imageUrl = card.querySelector(ebaySelectors.image)?.src ?? null;

  return { title, priceText, imageUrl };
}

export async function scanEbay(page: Page): Promise<DealNotification[]> {
  if (!globals.EBAY) return [];

  if (isAud) setStatus("Scanning eBay");
  else setStatus("Scanning eBay (USD)");

  const item = await getEbayQuery();
  if (!item) return [];

  const { foundName, foundPrice, foundImage } = await getEbayValues(page, item);
  if (!foundName || !foundPrice || !foundImage) return [];

  if (foundPrice > item.maxPrice) return [];

  if (await checkIfNew(foundName, SCANNER.EBAY)) {
    return [
      {
        kind: "deal",
        source: "ebay",
        title: `a ${foundName} priced at $${foundPrice} is available at ${item.url}`,
        url: item.url,
        price: foundPrice,
        imageUrl: foundImage,
        query: {
          type: "ebay",
          id: item.url,
        },
      },
    ];
  }
  return [];
}

export async function getEbayValues(page: Page, item: EbayQueryInput) {
  const gotoResult = await resultAsync(
    () =>
      page.goto(item.url, { waitUntil: "domcontentloaded", timeout: 10000 }),
    "eBay navigation failed",
  );
  if (gotoResult.isErr()) {
    console.warn(gotoResult.error.message);
    return { foundName: null, foundPrice: null, foundImage: null };
  }

  const selector = await selectorRace(
    page,
    ebaySelectors.resultsContainer,
    ebaySelectors.emptyResults,
  );
  if (!selector) return { foundName: null, foundPrice: null, foundImage: null };

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

            return { title, priceText, imageUrl };
          }),
        ebaySelectors,
      ),
    "eBay result card extraction failed",
  );
  if (rawItemsResult.isErr()) {
    console.warn(rawItemsResult.error.message);
    return { foundName: null, foundPrice: null, foundImage: null };
  }

  return parseEbayItem(
    rawItemsResult.value.find((rawItem) => rawItem.title) ?? null,
  );
}

let index = 0;
async function getEbayQuery() {
  const query = await db.ebay.findFirst({
    skip: index++,
  });
  if (query) {
    return query;
  }
  index = 1; //Will find the first query in the line below
  return await db.ebay.findFirst();
}
