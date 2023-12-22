import { Client, TextChannel } from "discord.js";
import puppeteer from "puppeteer";
import EbayQuery from "../../schema/ebayQuery.js";
import globals from "../../globals/Globals.js";
import client from "../../globals/DiscordJSClient.js";

let cursor = EbayQuery.find().cursor();
export async function scanEbay(page: puppeteer.Page) {
  if (!globals.EBAY || !globals.EBAY_CHANNEL_ID || !globals.EBAY_ROLE_ID)
    return;
  try {
    let item = await cursor.next();
    if (item === null) {
      cursor = EbayQuery.find().cursor();
      item = await cursor.next();
    }

    await page.goto(item.name);
    await page.waitForTimeout(Math.random() * 3000); //Waits before continuing. (Trying not to get IP banned)

    let foundName: string | undefined;
    let foundPrice: number | undefined;

    let result = await page.$("div[class='srp-river-results clearfix']"); //#s-item__wrapper
    if (result) {
      let resName = await result.$eval(
        "span[role='heading']",
        (res) => res.textContent
      );
      if (resName) foundName = resName;
      let resPrice = await result.$eval(
        "span[class='s-item__price']",
        (res) => res.textContent
      );
      if (resPrice) foundPrice = parseFloat(resPrice.replace(/[^0-9.-]+/g, ""));
    }

    if (
      foundName !== undefined &&
      foundPrice !== undefined &&
      foundName != item.lastItemFound &&
      foundPrice <= item.maxPrice
    ) {
      client.channels
        .fetch(globals.EBAY_CHANNEL_ID)
        .then((channel) => <TextChannel>channel)
        .then((channel) => {
          if (channel)
            channel.send(
              `<@&${globals.EBAY_ROLE_ID}> Please know that a ${foundName} priced at $${foundPrice} is available at ${item.name}`
            );
        })
        .catch((e) => console.error(e));
      if (foundName != undefined) {
        item.lastItemFound = foundName;
        item.save();
      }
    }
  } catch (e) {
    console.error(e);
  }
}
