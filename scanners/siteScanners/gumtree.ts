import puppeteer from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import sendToChannel from "../../functions/sendToChannel.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";
import { db, SCANNER } from "../../globals/PrismaClient.js";
import { checkIfNew } from "../../functions/handleItemUpdate.js";

export async function scanGumtree(page: puppeteer.Page) {
  if (
    !globals.GUMTREE ||
    !globals.GUMTREE_CHANNEL_ID ||
    !globals.GUMTREE_ROLE_ID
  )
    return;
  setStatus("Scanning Gumtree");

  let item = await getGumtreeQuery();

  await page.goto(item.url);

  let results = await page.$$(".user-ad-collection-new-design"); //#react-root > div > div.page > div > div.search-results-page__content > main > section > div

  //This uses a discrete solution instead of a selector race
  //Gumtree has rows of different categories (e.g. ads, out of area, ETC. This finds actual results, if it exists)
  let result;
  for (let i = 0; i < results.length; i++) {
    let tempRes = results[i];
    if (
      !(await tempRes.$(
        "div.user-ad-row-new-design__main-content > p.user-ad-row-new-design__title > span.user-ad-row-new-design__flag-top",
      ))
    ) {
      result = tempRes;
      break;
    }
  }
  if (!result) return;

  const foundName = await result.$eval(
    "div.user-ad-row-new-design__main-content > p.user-ad-row-new-design__title > span",
    (res) => res.textContent,
  );
  if (!foundName) return;

  let resPrice = await result.$eval(
    "div.user-ad-row-new-design__right-content > div:nth-child(1) > div > span.user-ad-price-new-design__price",
    (res) => res.textContent,
  );
  if (!resPrice) return;
  const foundPrice = resPrice.includes("Free")
    ? 0
    : parseFloat(resPrice.replace(/[^0-9.-]+/g, ""));

  if (await checkIfNew(foundName, SCANNER.GUMTREE))
    sendToChannel(
      globals.GUMTREE_CHANNEL_ID,
      `<@&${
        globals.GUMTREE_ROLE_ID
      }>${getNotificationPrelude()} a ${foundName} priced at $${foundPrice} is available at ${
        item.url
      }`,
    );
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
