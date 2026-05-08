import { Page } from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import { db, SCANNER } from "../../globals/PrismaClient.js";
import { checkIfNew } from "../../functions/handleItemUpdate.js";
import { resultAsync } from "../../functions/neverthrowUtils.js";
import type { DealNotification } from "../../deals/types.js";

interface GumtreeQueryInput {
  url: string;
}

export async function scanGumtree(page: Page): Promise<DealNotification[]> {
  if (!globals.GUMTREE) return [];
  setStatus("Scanning Gumtree");

  const item = await getGumtreeQuery();
  if (!item) return [];

  const result = await getGumtreeValues(page, item);
  if (!result) return [];

  //Skip if invalid price
  if (result.foundPrice > item.maxPrice) return [];

  if (await checkIfNew(result.foundName, SCANNER.GUMTREE)) {
    return [
      {
        kind: "deal",
        source: "gumtree",
        title: `a ${result.foundName} priced at $${result.foundPrice} is available at ${item.url}`,
        url: item.url,
        price: result.foundPrice,
        ...(result.foundImg && { imageUrl: result.foundImg }),
        query: {
          type: "gumtree",
          id: item.url,
        },
      },
    ];
  }
  return [];
}

export async function getGumtreeValues(page: Page, item: GumtreeQueryInput) {
  await page.goto(item.url);
  await page.setUserAgent(
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
  );
  const selectorResult = await resultAsync(
    () =>
      page.waitForSelector(
        "a[class='user-ad-row-new-design link link--base-color-inherit link--hover-color-none link--no-underline']",
        { timeout: 10000 },
      ),
    "Gumtree selector lookup failed",
  );
  if (selectorResult.isErr()) {
    console.warn(selectorResult.error.message);
    return null;
  }

  const result = await page.$(
    "a[class='user-ad-row-new-design link link--base-color-inherit link--hover-color-none link--no-underline']",
  );

  if (!result) return null;

  const foundName = await result.$eval(
    "div.user-ad-row-new-design__main-content > p.user-ad-row-new-design__title > span",
    (res) => res.textContent,
  );
  if (!foundName) return null;

  const resPrice = await result.$eval(
    "div.user-ad-row-new-design__right-content > div:nth-child(1) > div > span.user-ad-price-new-design__price",
    (res) => res.textContent,
  );
  if (!resPrice) return null;

  const foundPrice = resPrice.includes("Free")
    ? 0
    : parseFloat(resPrice.replace(/[^0-9.-]+/g, ""));

  const imgSelector = await result.$("img");
  const foundImg = await imgSelector?.evaluate((img) => img.src);

  return { foundName, foundPrice, foundImg };
}

let index = 0;
async function getGumtreeQuery() {
  const query = await db.gumtree.findFirst({
    skip: index++,
  });
  if (query) {
    return query;
  }
  index = 1; //Will find the first query in the line below
  return await db.gumtree.findFirst();
}
