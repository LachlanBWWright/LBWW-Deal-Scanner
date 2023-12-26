import { TextChannel } from "discord.js";
import puppeteer from "puppeteer";
import CashConvertersQuery from "../../schema/cashConvertersQuery.js";
import globals from "../../globals/Globals.js";
import client from "../../globals/DiscordJSClient.js";
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

  if (cashConvertersItem != item.lastItemFound) {
    sendToChannel(
      globals.CASH_CONVERTERS_CHANNEL_ID,
      `Please know that a ${cashConvertersItem} is available at ${item.name}`
    );
    if (cashConvertersItem != undefined) {
      item.lastItemFound = cashConvertersItem;
      item.save();
    }
  }
}
