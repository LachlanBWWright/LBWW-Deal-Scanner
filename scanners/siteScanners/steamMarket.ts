import { Client, TextChannel } from "discord.js";
import axios from "axios";
import SteamQuery from "../../schema/steamQuery.js";
import CsMarketItem from "../../schema/csMarketItem.js";
import puppeteer from "puppeteer";
import { CallbackError } from "mongoose";
import globals from "../../globals/Globals.js";
import client from "../../globals/DiscordJSClient.js";
import setStatus from "../../functions/setStatus.js";

//For general market queries and CS Items
let itemsFound = new Map<string, number>();
let queryCursor = SteamQuery.find().cursor();
let csCursor = CsMarketItem.find().cursor();

export async function scanSteamQuery() {
  if (!globals.CS_ITEMS || !globals.CS_CHANNEL_ID || !globals.CS_ROLE_ID)
    return;
  setStatus("Scanning the Steam Community Market");

  try {
    let item = await queryCursor.next();
    if (item === null) {
      queryCursor = SteamQuery.find().cursor();
      item = await queryCursor.next();
    }

    await sleep(3000);
    let res = await axios.get(`${item.name}`);
    if (res.status !== 200) return;
    let price: number = res.data.results[0].sell_price;
    for (let instance in res.data.results) {
      let thisPrice = parseFloat(res.data.results[instance].sell_price) / 100.0;
      if (thisPrice < price) price = thisPrice;
      if (price < item.maxPrice && price * 1.04 < item.lastPrice) {
        client.channels
          .fetch(globals.CS_CHANNEL_ID)
          .then((channel) => <TextChannel>channel)
          .then((channel) => {
            if (channel)
              channel.send(
                `<@&${globals.CS_ROLE_ID}> Please know that a ${res.data.results[instance].name} is available for $${price} USD at: ${item.displayUrl}`
              );
          })
          .catch((e) => console.error(e));
        break;
      }
    }
    if (price != item.lastPrice) {
      item.lastPrice = price;
      item.save((e: CallbackError) => console.error(e));
    }
  } catch (e) {
    console.error(e);
  }
}

//NOTE: This is depreciated currently due to increased ratelimits
export async function scanCs() {
  if (!globals.CS_ITEMS || !globals.CS_CHANNEL_ID || !globals.CS_ROLE_ID)
    return;

  try {
    let item = await csCursor.next();
    if (item === null) {
      csCursor = CsMarketItem.find().cursor();
      item = await csCursor.next();
    }

    await sleep(3000);
    let res = await axios.get(`${item.name}`);
    if (res.status !== 200) return;
    let i = 0;
    for (let skin in res.data.listinginfo) {
      let query = "https://api.csgofloat.com/?url="
        .concat(res.data.listinginfo[skin].asset.market_actions[0].link)
        .replace("%listingid%", res.data.listinginfo[skin].listingid)
        .replace("%assetid%", res.data.listinginfo[skin].asset.id);

      let price =
        (res.data.listinginfo[skin].converted_price_per_unit +
          res.data.listinginfo[skin].converted_fee_per_unit) /
        100.0;

      //Only calls the API if the skin isn't in the map, and the item is in the first 10
      if (!itemsFound.has(query) && i < 10) res = await axios.get(query);

      if (
        res.data.iteminfo.floatvalue < item.maxFloat &&
        price <= item.maxPrice
      ) {
        client.channels
          .fetch(globals.CS_CHANNEL_ID)
          .then((channel) => <TextChannel>channel)
          .then((channel) => {
            if (channel)
              channel.send(
                `<@&${globals.CS_ROLE_ID}> Please know that a ${res.data.iteminfo.full_item_name} with float ${res.data.iteminfo.floatvalue} is available for $${price} USD at: ${item.displayUrl}`
              );
          })
          .catch((e) => console.error(e));
      }

      if (i < 10) itemsFound.set(query, 20);
      //Puts the query into the map, or resets its TTL if the API was called for it
      else if (itemsFound.has(query)) itemsFound.set(query, 20); //Resets its TTL if it's already been called once
      i++;
    }

    //TODO: Refactor callback
    await axios
      .get(`${item.name}`)
      .then(async (res) => {
        let i = 0;
        for (let skin in res.data.listinginfo) {
          let query = "https://api.csgofloat.com/?url="
            .concat(res.data.listinginfo[skin].asset.market_actions[0].link)
            .replace("%listingid%", res.data.listinginfo[skin].listingid)
            .replace("%assetid%", res.data.listinginfo[skin].asset.id);

          let price =
            (res.data.listinginfo[skin].converted_price_per_unit +
              res.data.listinginfo[skin].converted_fee_per_unit) /
            100.0;

          //Only calls the API if the skin isn't in the map, and the item is in the first 10
          if (!itemsFound.has(query) && i < 10)
            await axios
              .get(query)
              .then((res) => {
                if (
                  res.data.iteminfo.floatvalue < item.maxFloat &&
                  price <= item.maxPrice
                ) {
                  client.channels
                    .fetch(globals.CS_CHANNEL_ID ?? "")
                    .then((channel) => <TextChannel>channel)
                    .then((channel) => {
                      if (channel)
                        channel.send(
                          `<@&${globals.CS_ROLE_ID}> Please know that a ${res.data.iteminfo.full_item_name} with float ${res.data.iteminfo.floatvalue} is available for $${price} USD at: ${item.displayUrl}`
                        );
                    })
                    .catch((e) => console.error(e));
                }
              })
              .catch((e) => console.error(e));
          if (i < 10) itemsFound.set(query, 20);
          //Puts the query into the map, or resets its TTL if the API was called for it
          else if (itemsFound.has(query)) itemsFound.set(query, 20); //Resets its TTL if it's already been called once
          i++;
        }
      })
      .catch((e) => console.error(e));

    //Decrement the TTL in the map
    for (let [key, value] of itemsFound) {
      value--;
      if (value <= 0) itemsFound.delete(key);
    }
  } catch (e) {
    console.error(e);
  }
}

export async function createQuery(
  oldQuery: string,
  maxPrice: number
): Promise<[boolean, string]> {
  let browser = await puppeteer.launch({
    headless: true,
    args: ["--no-sandbox"],
  });
  const page = await browser.newPage();
  let response: [boolean, string];
  response = [false, ""];
  try {
    if (!oldQuery.includes("https://steamcommunity.com/market/search")) {
      response = [false, ""];
    }

    let responseUrl = "";
    let query = new URL(oldQuery);
    page.goto(query.toString());

    //New eventlistener replacement
    await page.waitForResponse((response) => {
      if (
        response
          .url()
          .includes("https://steamcommunity.com/market/search/render/?query")
      ) {
        new URL(query);
        let steamQuery = new SteamQuery({
          name: response.url().toString().concat("&norender=1"),
          displayUrl: oldQuery,
          maxPrice: maxPrice,
        });
        steamQuery.save((e) => console.error(e));
        responseUrl = response.url().toString().concat("&norender=1"); //Makes it JSON instead of html
        return true;
      }
      return false;
    });
    response = [true, responseUrl];
  } catch (e) {
    console.error(e);
    response = [false, ""];
  } finally {
    await page.close();
    await browser.disconnect();
    //await browser.close();
    return response;
  }
}
export async function createCs(
  oldQuery: string,
  maxPrice: number,
  maxFloat: number
) {
  //Init. Example: https://steamcommunity.com/market/listings/730/M4A1-S%20%7C%20Chantico%27s%20Fire%20%28Field-Tested%29
  //Conv. example: https://steamcommunity.com/market/listings/730/M4A1-S%20%7C%20Chantico%27s%20Fire%20%28Field-Tested%29/render/?query=&start=0&count=10&country=AU&language=english&currency=1
  try {
    if (oldQuery.includes("https://steamcommunity.com/market/listings/730/")) {
      let search = new URL(
        oldQuery
          .concat(
            "/render/?query=&start=0&count=20&country=AU&language=english&currency=1"
          )
          .trim()
      ).toString();
      let csMarketItem = new CsMarketItem({
        name: search,
        displayUrl: oldQuery,
        maxPrice: maxPrice,
        maxFloat: maxFloat,
      });

      await csMarketItem.save((e) => console.error(e));
      return search;
    }
    return "";
  } catch (e) {
    console.error(e);
    return "";
  }
}

function sleep(ms: number) {
  return new Promise((res) => setTimeout(res, ms));
}
