import { Page } from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import selectorRace from "../../functions/selectorRace.js";
import sendToChannel from "../../functions/sendToChannel.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";
import { db, SCANNER } from "../../globals/PrismaClient.js";
import { checkIfNew } from "../../functions/handleItemUpdate.js";
import { Ebay } from "@prisma/client";

export async function scanEbay(page: Page) {
  if (!globals.EBAY || !globals.EBAY_CHANNEL_ID || !globals.EBAY_ROLE_ID)
    return;
  setStatus("Scanning eBay");

  const item = await getEbayQuery();
  await getEbayValues(page, item);

  const { foundName, foundPrice, foundImage } = await getEbayValues(page, item);

  if (foundPrice > item.maxPrice) return;

  if (await checkIfNew(foundName, SCANNER.EBAY)) {
    sendToChannel(
      globals.EBAY_CHANNEL_ID,
      `<@&${
        globals.EBAY_ROLE_ID
      }> ${getNotificationPrelude()} a ${foundName} priced at $${foundPrice} is available at ${
        item.url
      }`,
      { files: [foundImage] },
    );
  }
}

export async function getEbayValues(page: Page, item: Ebay) {
  await page.goto(item.url);

  const selector = await selectorRace(
    page,
    "div[class='srp-river-results clearfix']",
    ".srp-save-null-search__heading",
  );
  if (!selector) throw new Error("Missing eBay selector");

  const result = await selector.$("li[class='s-item s-item__pl-on-bottom']");
  if (!result) throw new Error("Missing eBay");

  const foundName = await result.$eval('span[role="heading"]', (res) => {
    if (res.textContent?.startsWith("New listing"))
      return res.textContent.replace("New listing", "");
    return res.textContent;
  });
  const foundPrice = await result.$eval("span[class='s-item__price']", (res) =>
    res.textContent
      ? parseFloat(res.textContent.replace(/[^0-9.-]+/g, ""))
      : null,
  );

  const foundImgContainer = await result.$(
    "div[class='s-item__image-wrapper image-treatment']",
  );

  const foundImage = await foundImgContainer?.$eval("img", (img) => img.src);

  if (!foundName || !foundPrice || !foundImage)
    throw new Error("Could not find name, price, or img");

  return { foundName, foundPrice, foundImage };
}

let index = 0;
async function getEbayQuery() {
  let query = await db.ebay.findFirst({
    skip: index++,
  });
  if (query) {
    index++;
    return query;
  }
  index = 1; //Will find the first query in the line below
  return await db.ebay.findFirstOrThrow();
}
