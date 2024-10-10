import { Page } from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import sendToChannel from "../../functions/sendToChannel.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";
import { db, SCANNER } from "../../globals/PrismaClient.js";
import { checkIfNew } from "../../functions/handleItemUpdate.js";
import { Gumtree } from "@prisma/client";

export async function scanGumtree(page: Page) {
  if (
    !globals.GUMTREE ||
    !globals.GUMTREE_CHANNEL_ID ||
    !globals.GUMTREE_ROLE_ID
  )
    return;
  setStatus("Scanning Gumtree");

  let item = await getGumtreeQuery();

  const result = await getGumtreeValues(page, item);
  if (!result) return;

  //Skip if invalid price
  if (result.foundPrice > item.maxPrice) return;

  if (await checkIfNew(result.foundName, SCANNER.GUMTREE))
    sendToChannel(
      globals.GUMTREE_CHANNEL_ID,
      `<@&${globals.GUMTREE_ROLE_ID}>${getNotificationPrelude()} a ${
        result.foundName
      } priced at $${result.foundPrice} is available at ${item.url}`,
      { files: [result.foundImg] },
    );
}

export async function getGumtreeValues(page: Page, item: Gumtree) {
  await page.goto(item.url);
  await page.setUserAgent(
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
  );
  await page.waitForSelector(
    "a[class='user-ad-row-new-design link link--base-color-inherit link--hover-color-none link--no-underline']",
  );

  let result = await page.$(
    "a[class='user-ad-row-new-design link link--base-color-inherit link--hover-color-none link--no-underline']",
  ); //#react-root > div > div.page > div > div.search-results-page__content > main > section > div

  if (!result) throw new Error("No ressult");

  //This uses a discrete solution instead of a selector race
  //Gumtree has rows of different categories (e.g. ads, out of area, ETC. This finds actual results, if it exists)

  const foundName = await result.$eval(
    "div.user-ad-row-new-design__main-content > p.user-ad-row-new-design__title > span",
    (res) => res.textContent,
  );
  if (!foundName) throw new Error();

  let resPrice = await result.$eval(
    "div.user-ad-row-new-design__right-content > div:nth-child(1) > div > span.user-ad-price-new-design__price",
    (res) => res.textContent,
  );
  if (!resPrice) throw new Error();
  const foundPrice = resPrice.includes("Free")
    ? 0
    : parseFloat(resPrice.replace(/[^0-9.-]+/g, ""));

  const foundImg = await result.$eval("img", (img) => img.src);

  return { foundName, foundPrice, foundImg };
}

let index = 0;
async function getGumtreeQuery() {
  let query = await db.gumtree.findFirst({
    skip: index++,
  });
  if (query) {
    index++;
    return query;
  }
  index = 1; //Will find the first query in the line below
  return await db.gumtree.findFirstOrThrow();
}
