import { Page } from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import selectorRace from "../../functions/selectorRace.js";
import sendToChannel from "../../functions/sendToChannel.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";
import { db, SCANNER } from "../../globals/PrismaClient.js";
import { checkIfNew } from "../../functions/handleItemUpdate.js";
import { Ebay } from "@prisma/client";

let isAud = true;

export async function scanEbay(page: Page) {
  if (!globals.EBAY || !globals.EBAY_CHANNEL_ID || !globals.EBAY_ROLE_ID)
    return;

  if (isAud) setStatus("Scanning eBay");
  else setStatus("Scanning eBay (USD)");

  const item = await getEbayQuery();
  if (!item) return;

  const { foundName, foundPrice, foundImage } = await getEbayValues(page, item);
  if (!foundName || !foundPrice || !foundImage) return;

  if (foundPrice > item.maxPrice) return;

  if (await checkIfNew(foundName, SCANNER.EBAY)) {
    await sendToChannel(
      globals.EBAY_CHANNEL_ID,
      `<@&${
        globals.EBAY_ROLE_ID
      }> ${getNotificationPrelude()} a ${foundName} priced at $${foundPrice} is available at ${
        item.url
      }`,
      {
        files: [foundImage],
        queryId: item.url,
        queryType: "ebay",
      },
    );
  }
}

export async function getEbayValues(page: Page, item: Ebay) {
  try {
    await page.goto(item.url, { waitUntil: "domcontentloaded", timeout: 10000 });
  } catch {
    return { foundName: null, foundPrice: null, foundImage: null };
  }

  console.log("test 1");
  console.log(page);

  const test = await page.title();
  console.log(test);

  const selector = await selectorRace(
    page,
    "div[class='srp-river']",
    ".srp-save-null-search__heading",
  );
  if (!selector) return { foundName: null, foundPrice: null, foundImage: null };

  console.log("test 2");

  const result = await selector.$(
    `div[class="su-card-container su-card-container--horizontal"]`,
  );
  if (!result) return { foundName: null, foundPrice: null, foundImage: null };

  const foundName = await result.$eval('div[role="heading"]', (res) => {
    if (res.textContent?.startsWith("New listing"))
      return res.textContent.replace("New listing", "");
    return res.textContent;
  });
  const foundPrice = await result.$eval(
    `span[class='su-styled-text primary bold large-1 s-card__price']`,
    (res) => {
      if (res.textContent?.startsWith("AU $")) {
        isAud = true;
        return res.textContent
          ? parseFloat(res.textContent.replace(/[^0-9.-]+/g, ""))
          : null;
      }
      isAud = false;
      return res.textContent
        ? parseFloat(res.textContent.replace(/[^0-9.-]+/g, "")) * 1.55 //Hack currency conversion
        : null;
    },
  );

  const foundImgContainer = await result.$("a[class='image-treatment']");

  const foundImage = await foundImgContainer?.$eval("img", (img) => img.src);

  console.log(foundName, foundPrice, foundImage);
  if (!foundName || !foundPrice || !foundImage) {
    return { foundName: null, foundPrice: null, foundImage: null };
  }

  return { foundName, foundPrice, foundImage };
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
