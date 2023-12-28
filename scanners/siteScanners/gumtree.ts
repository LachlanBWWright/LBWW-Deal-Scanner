import puppeteer from "puppeteer";
import GumtreeQuery from "../../schema/gumtreeQuery.js";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import sendToChannel from "../../functions/sendToChannel.js";

let cursor = GumtreeQuery.find().cursor();
let recentlyFound = new Map<string, Map<string, number>>();

export async function scanGumtree(page: puppeteer.Page) {
  if (
    !globals.GUMTREE ||
    !globals.GUMTREE_CHANNEL_ID ||
    !globals.GUMTREE_ROLE_ID
  )
    return;
  setStatus("Scanning Gumtree");

  let item = await cursor.next();
  if (item === null) {
    cursor = GumtreeQuery.find().cursor();
    item = await cursor.next();
  }

  await page.goto(item.name);

  let results = await page.$$(".user-ad-collection-new-design"); //#react-root > div > div.page > div > div.search-results-page__content > main > section > div

  //This uses a discrete solution instead of a selector race
  //Gumtree has rows of different categories (e.g. ads, out of area, ETC. This finds actual results, if it exists)
  let result;
  for (let i = 0; i < results.length; i++) {
    let tempRes = results[i];
    if (
      !(await tempRes.$(
        "div.user-ad-row-new-design__main-content > p.user-ad-row-new-design__title > span.user-ad-row-new-design__flag-top"
      ))
    ) {
      result = tempRes;
      break;
    }
  }
  if (!result) return;

  const foundName = await result.$eval(
    "div.user-ad-row-new-design__main-content > p.user-ad-row-new-design__title > span",
    (res) => res.textContent
  );
  if (!foundName) return;

  let resPrice = await result.$eval(
    "div.user-ad-row-new-design__right-content > div:nth-child(1) > div > span.user-ad-price-new-design__price",
    (res) => res.textContent
  );
  if (!resPrice) return;
  const foundPrice = resPrice.includes("Free")
    ? 0
    : parseFloat(resPrice.replace(/[^0-9.-]+/g, ""));

  //Map stuff
  let foundRecently = true;
  let map =
    recentlyFound.get(item.name) ??
    recentlyFound.set(item.name, new Map()).get(item.name); //Expands map if item does not exist in it.
  if (map !== undefined && foundName !== undefined) {
    if (map.get(foundName) === undefined) foundRecently = false; //Block notification if date was already found
    map.set(foundName, Date.now()); //MS pased since 1970, lower = older

    if (map.size > 10) {
      //Purges the youngest from the map
      let toDelete = "";
      let oldest = 0;
      map.forEach((val, key) => {
        if (val < oldest || oldest === 0) {
          toDelete = key;
          oldest = val;
        }
      });
      map.delete(toDelete);
    }
  }

  if (
    foundName === item.lastItemFound ||
    foundPrice > item.maxPrice ||
    foundRecently
  )
    return;

  sendToChannel(
    globals.GUMTREE_CHANNEL_ID,
    `<@&${globals.GUMTREE_ROLE_ID}> Please know that a ${foundName} priced at $${foundPrice} is available at ${item.name}`
  );

  if (foundName != undefined) {
    item.lastItemFound = foundName;
    item.save();
  }
}
