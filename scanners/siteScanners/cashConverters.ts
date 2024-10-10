import { Page } from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import selectorRace from "../../functions/selectorRace.js";
import sendToChannel from "../../functions/sendToChannel.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";
import { db, SCANNER } from "../../globals/PrismaClient.js";
import { formatPrice } from "../../functions/formatPrice.js";
import { checkIfNew } from "../../functions/handleItemUpdate.js";
import { CashConverters } from "@prisma/client";

export async function scanCashConverters(page: Page) {
  if (
    !globals.CASH_CONVERTERS ||
    !globals.CASH_CONVERTERS_CHANNEL_ID ||
    !globals.CASH_CONVERTERS_ROLE_ID
  )
    return;
  setStatus("Scanning Cash Converters");

  const item = await getCashQuery();
  const foundItem = await getCashConvertersValues(page, item);
  if (!foundItem) return;

  if (await checkIfNew(foundItem.itemName, SCANNER.CASH_CONVERTERS)) {
    sendToChannel(
      globals.CASH_CONVERTERS_CHANNEL_ID,
      `<@&${globals.CASH_CONVERTERS_ROLE_ID}> ${getNotificationPrelude()} a ${
        foundItem.itemName
      } for ${formatPrice(foundItem.totalPrice)} is available at ${
        foundItem.itemName
      }`,
    );
  }
}

export async function getCashConvertersValues(
  page: Page,
  item: CashConverters,
) {
  await page.goto(item.url);
  const selector = await selectorRace(
    page,
    "div[class='product-item']",
    ".no-search-results__text",
  );
  if (!selector) return;
  const itemName = await selector.evaluate((el) => el.textContent);
  if (!itemName) return;

  const price = await selector.$eval(".product-item__price", (selector) =>
    //Slice removes the '$'
    selector.textContent ? parseFloat(selector.textContent.slice(1)) : null,
  );
  if (!price) return;

  const shipping = await selector.$eval(".product-item__postage", (selector) =>
    //Slice removes the '+ $'
    selector.textContent ? parseFloat(selector.textContent.slice(3)) : null,
  );
  if (shipping === null) return;

  const image = await selector.$eval("img", (img) => img.src);

  const totalPrice = price + (isNaN(shipping) ? 0 : shipping);

  return { itemName, totalPrice, image };
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
