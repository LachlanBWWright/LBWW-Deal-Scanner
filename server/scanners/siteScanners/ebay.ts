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
    return nameText.replace("New listing", "");
  }
  return nameText;
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
    "div[class='srp-river']",
    ".srp-save-null-search__heading",
  );
  if (!selector) return { foundName: null, foundPrice: null, foundImage: null };

  const result = await selector.$(
    `div[class="su-card-container su-card-container--horizontal"]`,
  );
  if (!result) return { foundName: null, foundPrice: null, foundImage: null };

  const foundName = await result.$eval('div[role="heading"]', (res) => res.textContent);
  const priceText = await result.$eval(
    `span[class='su-styled-text primary bold large-1 s-card__price']`,
    (res) => res.textContent,
  );
  const foundPrice = parseEbayPrice(priceText ?? null);
  const parsedName = parseEbayName(foundName ?? null);

  const foundImgContainer = await result.$("a[class='image-treatment']");
  const foundImage = await foundImgContainer?.$eval("img", (img) => img.src);

  if (!parsedName || !foundPrice || !foundImage) {
    return { foundName: null, foundPrice: null, foundImage: null };
  }

  return { foundName: parsedName, foundPrice, foundImage };
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
