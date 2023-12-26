import { ChannelType, TextChannel } from "discord.js";
import puppeteer from "puppeteer";
import EbayQuery from "../../schema/ebayQuery.js";
import globals from "../../globals/Globals.js";
import client from "../../globals/DiscordJSClient.js";
import setStatus from "../../functions/setStatus.js";
import selectorRace from "../../functions/selectorRace.js";
import sendToChannel from "../../functions/sendToChannel.js";

let cursor = EbayQuery.find().cursor();
export async function scanEbay(page: puppeteer.Page) {
  if (!globals.EBAY || !globals.EBAY_CHANNEL_ID || !globals.EBAY_ROLE_ID)
    return;
  setStatus("Scanning eBay");

  let item = await cursor.next();
  if (item === null) {
    cursor = EbayQuery.find().cursor();
    item = await cursor.next();
  }
  await page.goto(item.name);

  const selector = await selectorRace(
    page,
    "div[class='srp-river-results clearfix']",
    ".srp-save-null-search__heading"
  );
  if (!selector) return;

  const foundName = await selector.$eval(
    'span[role="heading"]',
    (res) => res.textContent
  );
  const foundPrice = await selector.$eval(
    "span[class='s-item__price']",
    (res) =>
      res.textContent
        ? parseFloat(res.textContent.replace(/[^0-9.-]+/g, ""))
        : null
  );
  if (!foundName || !foundPrice)
    throw new Error("Could not find name or price");
  if (foundName === item.lastItemFound || foundPrice > item.maxPrice) return;

  sendToChannel(
    globals.EBAY_CHANNEL_ID,
    `Please know that a ${foundName} priced at $${foundPrice} is available at ${item.name}`
  );

  if (foundName != undefined) {
    item.lastItemFound = foundName;
    await item.save();
  }
}
