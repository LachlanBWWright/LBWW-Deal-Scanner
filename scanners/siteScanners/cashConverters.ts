import puppeteer from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import selectorRace from "../../functions/selectorRace.js";
import sendToChannel from "../../functions/sendToChannel.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";
import { db, SCANNER } from "../../globals/PrismaClient.js";
import { formatPrice } from "../../functions/formatPrice.js";
import { checkIfNew } from "../../functions/handleItemUpdate.js";

export async function scanCashConverters(page: puppeteer.Page) {
  if (
    !globals.CASH_CONVERTERS ||
    !globals.CASH_CONVERTERS_CHANNEL_ID ||
    !globals.CASH_CONVERTERS_ROLE_ID
  )
    return;
  setStatus("Scanning Cash Converters");

  const query = await getCashQuery();

  await page.goto(query.url);
  const selector = await selectorRace(
    page,
    ".product-item__title__description",
    ".no-search-results__text",
  );
  if (!selector) return;

  const itemName = await selector.evaluate((el) => el.textContent);
  if (!itemName) return;

  const price = await page.$eval(".product-item__price", (selector) =>
    //Slice removes the '$'
    selector.textContent ? parseFloat(selector.textContent.slice(1)) : null,
  );
  if (!price) return;

  const shipping = await page.$eval(".product-item__postage", (selector) =>
    //Slice removes the '+ $'
    selector.textContent ? parseFloat(selector.textContent.slice(3)) : null,
  );
  if (shipping === null) return;

  const totalPrice = price + (isNaN(shipping) ? 0 : shipping);

  if (await checkIfNew(itemName, SCANNER.CASH_CONVERTERS)) {
    sendToChannel(
      globals.CASH_CONVERTERS_CHANNEL_ID,
      `<@&${
        globals.CASH_CONVERTERS_ROLE_ID
      }> ${getNotificationPrelude()} a ${itemName} for ${formatPrice(
        totalPrice,
      )} is available at ${itemName}`,
    );
  }
}

let index = 0;
async function getCashQuery() {
  let query = await db.cashConverters.findFirst({
    skip: index++,
  });
  if (query) {
    index++;
    return query;
  }
  index = 1; //Will find the first query in the line below
  return await db.cashConverters.findFirstOrThrow();
}
