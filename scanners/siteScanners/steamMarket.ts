import axios from "axios";
import { ResultAsync } from "neverthrow";
import { Page } from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import sendToChannel from "../../functions/sendToChannel.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";
import { db } from "../../globals/PrismaClient.js";

//For general market queries and CS Items
const itemsFound = new Map<string, number>();

export async function scanSteamQuery() {
  if (
    !globals.STEAM_QUERY ||
    !globals.STEAM_QUERY_CHANNEL_ID ||
    !globals.STEAM_QUERY_ROLE_ID
  )
    return;
  setStatus("Scanning the Steam Community Market");

  try {
    const item = await getSteamQuery();
    if (!item) return;

    await sleep(3000);

    const results = await getQueryResults(item.name);
    if (results.length === 0) return;
    const result = results[0];
    const price = parseFloat(result.sell_price) / 100.0;
    if (price < item.maxPrice && price * 1.04 < item.lastPrice) {
      await sendToChannel(
        globals.STEAM_QUERY_CHANNEL_ID,
        `<@&${globals.STEAM_QUERY_ROLE_ID}> ${getNotificationPrelude()} a ${
          result.name
        } is available for $${price} USD at: ${item.displayUrl}`,
        {
          queryId: item.name,
          queryType: 'steamMarket'
        },
      );
    }

    if (price != item.lastPrice) {
      await db.steamMarket.update({
        where: { name: item.name },
        data: {
          lastPrice: price,
        },
      });
    }
  } catch (e) {
    console.error(e);
  }
}

export async function getQueryResults(url: string) {
  return ResultAsync.fromPromise(
    axios.get(url),
    () => new Error("Steam query failed"),
  )
    .andThen((response) =>
      response.status === 200
        ? ResultAsync.fromSafePromise(Promise.resolve(response.data.results))
        : ResultAsync.fromSafePromise(Promise.resolve([])),
    )
    .match(
      (results) => results,
      () => [],
    );
}

//NOTE: This is depreciated currently due to increased ratelimits
export async function scanCs() {
  if (!globals.CS_ITEMS || !globals.CS_CHANNEL_ID || !globals.CS_ROLE_ID)
    return;

  try {
    const item = await getCsMarketQuery();
    if (!item) return;

    await sleep(3000);
    let res = await axios.get(`${item.url}`);
    if (res.status !== 200) return;
    let i = 0;
    for (const skin in res.data.listinginfo) {
      const query = "https://api.csgofloat.com/?url="
        .concat(res.data.listinginfo[skin].asset.market_actions[0].link)
        .replace("%listingid%", res.data.listinginfo[skin].listingid)
        .replace("%assetid%", res.data.listinginfo[skin].asset.id);

      const price =
        (res.data.listinginfo[skin].converted_price_per_unit +
          res.data.listinginfo[skin].converted_fee_per_unit) /
        100.0;

      //Only calls the API if the skin isn't in the map, and the item is in the first 10
      if (!itemsFound.has(query) && i < 10) res = await axios.get(query);

      if (
        res.data.iteminfo.floatvalue < item.maxFloat &&
        price <= item.maxPrice
      ) {
        await sendToChannel(
          globals.CS_CHANNEL_ID,
          `<@&${globals.CS_ROLE_ID}> ${getNotificationPrelude()} a ${
            res.data.iteminfo.full_item_name
          } with float ${
            res.data.iteminfo.floatvalue
          } is available for $${price} USD at: ${item.displayUrl}`,
          {
            queryId: item.url,
            queryType: 'csMarket'
          },
        );
      }

      if (i < 10) itemsFound.set(query, 20);
      //Puts the query into the map, or resets its TTL if the API was called for it
      else if (itemsFound.has(query)) itemsFound.set(query, 20); //Resets its TTL if it's already been called once
      i++;
    }

    //Decrement the TTL in the map
  for (const [key, value] of itemsFound) {
    const newValue = value - 1;
    if (newValue <= 0) itemsFound.delete(key);
    else itemsFound.set(key, newValue);
    }
  } catch (e) {
    console.error(e);
  }
}

export async function getCsQueryString(page: Page, oldQuery: string) {
  if (!oldQuery.includes("https://steamcommunity.com/market/search")) {
    return "";
  }

  //Deliberately not awaited, as response will resolve (and become unavailable) before waitForResponse
  void page.goto(new URL(oldQuery).toString());

  //Promise resolves when response handling function returns true
  const response = await page.waitForResponse(
    (response) =>
      response
        .url()
        .includes("https://steamcommunity.com/market/search/render/?query"),

    { timeout: 10000 },
  );
  await page.waitForNetworkIdle();
  return response.url().toString().concat("&norender=1"); //Makes it JSON instead of html
}
export async function createCs(
  oldQuery: string,
  maxPrice: number,
  maxFloat: number,
  dmOnly = false,
) {
  //Init. Example: https://steamcommunity.com/market/listings/730/M4A1-S%20%7C%20Chantico%27s%20Fire%20%28Field-Tested%29
  //Conv. example: https://steamcommunity.com/market/listings/730/M4A1-S%20%7C%20Chantico%27s%20Fire%20%28Field-Tested%29/render/?query=&start=0&count=10&country=AU&language=english&currency=1
  try {
    if (oldQuery.includes("https://steamcommunity.com/market/listings/730/")) {
      const search = new URL(
        oldQuery
          .concat(
            "/render/?query=&start=0&count=20&country=AU&language=english&currency=1",
          )
          .trim(),
      ).toString();

      await db.query.create({
        data: {
          dmOnly,
          csMarket: {
            create: {
              url: search,
              displayUrl: oldQuery,
              maxPrice: maxPrice,
              maxFloat: maxFloat,
              lastPrice: 0,
            },
          },
        },
      });

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

let steamQueryIndex = 0;
async function getSteamQuery() {
  const query = await db.steamMarket.findFirst({
    skip: steamQueryIndex++,
  });
  if (query) {
    return query;
  }
  steamQueryIndex = 1; //Will find the first query in the line below
  return await db.steamMarket.findFirst();
}

let csMarketIndex = 0;
async function getCsMarketQuery() {
  const query = await db.csMarket.findFirst({
    skip: csMarketIndex++,
  });
  if (query) {
    return query;
  }
  csMarketIndex = 1; //Will find the first query in the line below
  return await db.csMarket.findFirst();
}
