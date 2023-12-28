import puppeteer from "puppeteer";
import CashConvertersQuery from "../../schema/cashConvertersQuery.js";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import selectorRace from "../../functions/selectorRace.js";
import sendToChannel from "../../functions/sendToChannel.js";

let cursor = CashConvertersQuery.find().cursor();

export async function scanCashConverters(page: puppeteer.Page) {
  if (
    !globals.CASH_CONVERTERS ||
    !globals.CASH_CONVERTERS_CHANNEL_ID ||
    !globals.CASH_CONVERTERS_ROLE_ID
  )
    return;
  setStatus("Scanning Cash Converters");

  let item = await cursor.next();
  if (item === null) {
    cursor = CashConvertersQuery.find().cursor();
    item = await cursor.next();
  }
  await page.goto(item.name);
  let selector = await selectorRace(
    page,
    ".product-item__title__description",
    ".no-search-results__text"
  );

  if (!selector) return;
  let cashConvertersItem = await selector.evaluate((el) => el.textContent);

  const price = await page.$eval(".product-item__price", (selector) =>
    //Slice removes the '$'
    selector.textContent ? parseFloat(selector.textContent.slice(1)) : null
  );
  if (!price) return;

  const shipping = await page.$eval(".product-item__postage", (selector) =>
    //Slice removes the '+ $'
    selector.textContent ? parseFloat(selector.textContent.slice(3)) : null
  );
  if (shipping === null) return;

  const totalPrice = price + (isNaN(shipping) ? 0 : shipping);

  if (cashConvertersItem != item.lastItemFound) {
    sendToChannel(
      globals.CASH_CONVERTERS_CHANNEL_ID,
      `<@&${globals.CASH_CONVERTERS_ROLE_ID}> Please know that a ${cashConvertersItem} for $${totalPrice} is available at ${item.name}`
    );
    if (cashConvertersItem != undefined) {
      item.lastItemFound = cashConvertersItem;
      item.save();
    }
  }
}
