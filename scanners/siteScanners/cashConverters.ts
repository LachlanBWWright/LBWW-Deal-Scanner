import { Client, TextChannel } from "discord.js";
import puppeteer from "puppeteer";
import CashConvertersQuery from "../../schema/cashConvertersQuery.js";
import globals from "../../globals/Globals.js";
import client from "../../globals/DiscordJSClient.js";

let cursor = CashConvertersQuery.find().cursor();

export async function scanCashConverters(page: puppeteer.Page) {
  if (
    !globals.CASH_CONVERTERS ||
    !globals.CASH_CONVERTERS_CHANNEL_ID ||
    !globals.CASH_CONVERTERS_ROLE_ID
  )
    return;

  let item = await cursor.next();
  if (item === null) {
    cursor = CashConvertersQuery.find().cursor();
    item = await cursor.next();
  }

  await page.goto(item.name);
  await page.waitForTimeout(Math.random() * 3000); //Waits before continuing. (Trying not to get IP banned)
  let selector = await page.waitForSelector(
    ".product-item__title__description"
  );
  let cashConvertersItem = await selector?.evaluate((el) => el.textContent);
  if (cashConvertersItem != item.lastItemFound) {
    client.channels
      .fetch(globals.CASH_CONVERTERS_CHANNEL_ID)
      .then((channel) => <TextChannel>channel)
      .then((channel) => {
        if (channel)
          channel.send(
            `<@&${globals.CASH_CONVERTERS_ROLE_ID}> Please know that a ${cashConvertersItem} is available at  ${item.name}`
          );
      })
      .catch((e) => console.error(e));
    if (cashConvertersItem != undefined) {
      item.lastItemFound = cashConvertersItem;
      item.save();
    }
  }
}
