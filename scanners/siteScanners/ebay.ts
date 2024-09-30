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
  setStatus("Scanning eBay");

  const item = await getEbayQuery();
  checkEbayQuery(page, item);
}

export async function checkEbayQuery(page: Page, item: Ebay) {
  if (!globals.EBAY || !globals.EBAY_CHANNEL_ID || !globals.EBAY_ROLE_ID)
    return;
  await page.goto(item.url);

  const selector = await selectorRace(
    page,
    "div[class='srp-river-results clearfix']",
    ".srp-save-null-search__heading",
  );
  if (!selector) return;
  const foundName = await selector.$eval(
    'span[role="heading"]',
    (res) => res.textContent,
  );
  const foundPrice = await selector.$eval(
    "span[class='s-item__price']",
    (res) =>
      res.textContent
        ? parseFloat(res.textContent.replace(/[^0-9.-]+/g, ""))
        : null,
  );
  if (!foundName || !foundPrice)
    throw new Error("Could not find name or price");

  if (foundPrice > item.maxPrice) return;

  if (await checkIfNew(foundName, SCANNER.EBAY)) {
    sendToChannel(
      globals.EBAY_CHANNEL_ID,
      `<@&${
        globals.EBAY_ROLE_ID
      }> ${getNotificationPrelude()} a ${foundName} priced at $${foundPrice} is available at ${
        item.url
      }`,
    );
  }
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
