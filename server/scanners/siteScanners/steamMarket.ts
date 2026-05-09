import { Page } from "puppeteer";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import { db } from "../../globals/PrismaClient.js";
import { resultAsync } from "../../functions/neverthrowUtils.js";
import {
  fetchSteamMarketResults,
  fetchSteamCsMarketListing,
  fetchSteamItemInfo,
} from "../../functions/steamMarketFetcher.js";
import type { DealNotification } from "../../deals/types.js";
import type { SteamCsMarketResponse } from "../../functions/apiValidators.js";

type CsMarketQuery = Awaited<ReturnType<typeof getCsMarketQuery>>;

//For general market queries and CS Items
const itemsFound = new Map<string, number>();

export async function scanSteamQuery(): Promise<DealNotification[]> {
  if (!globals.STEAM_QUERY) return [];
  setStatus("Scanning the Steam Community Market");

  const notifications: DealNotification[] = [];

  const scanResult = await resultAsync(async () => {
    const item = await getSteamQuery();
    if (!item) return;

    await sleep(3000);

    const results = await getQueryResults(item.name);
    if (results.length === 0) return;
    const result = results[0];
    const price = result.sell_price / 100.0;
    if (price < item.maxPrice && price * 1.04 < item.lastPrice) {
      notifications.push({
        kind: "deal",
        source: "steamMarket",
        title: `a ${result.name} is available for $${price} USD at: ${item.displayUrl}`,
        url: item.displayUrl,
        price,
        query: {
          type: "steamMarket",
          id: item.name,
        },
      });
    }

    if (price !== item.lastPrice) {
      await db.steamMarket.update({
        where: { name: item.name },
        data: {
          lastPrice: price,
        },
      });
    }
  }, "Steam query scan failed");

  if (scanResult.isErr()) {
    console.error(scanResult.error.message);
  }

  return notifications;
}

export async function getQueryResults(url: string) {
  const result = await fetchSteamMarketResults(url);

  if (result.isErr()) {
    console.warn("Failed to fetch Steam market results:", result.error.message);
    return [];
  }

  return result.value;
}

//NOTE: This is depreciated currently due to increased ratelimits
async function processCsMarketListings(
  item: NonNullable<CsMarketQuery>,
  listingData: SteamCsMarketResponse,
  notifications: DealNotification[],
) {
  let i = 0;
  for (const skin in listingData.listinginfo) {
    const listing = listingData.listinginfo[skin];
    const query = "https://api.csgofloat.com/?url="
      .concat(listing.asset.market_actions[0].link)
      .replace("%listingid%", listing.listingid)
      .replace("%assetid%", listing.asset.id);

    const price =
      (listing.converted_price_per_unit + listing.converted_fee_per_unit) /
      100.0;

    if (!itemsFound.has(query) && i < 10) {
      const itemInfoResult = await fetchSteamItemInfo(query);
      if (itemInfoResult.isErr()) {
        console.warn(
          "Failed to fetch item info:",
          itemInfoResult.error.message,
        );
        i++;
        continue;
      }

      const itemInfo = itemInfoResult.value;
      if (
        itemInfo.iteminfo.floatvalue < item.maxFloat &&
        price <= item.maxPrice
      ) {
        notifications.push({
          kind: "deal",
          source: "steamMarket",
          title: `a ${itemInfo.iteminfo.full_item_name} with float ${itemInfo.iteminfo.floatvalue} is available for $${price} USD at: ${item.displayUrl}`,
          url: item.displayUrl,
          price,
          query: {
            type: "csMarket",
            id: item.url,
          },
        });
      }
    }

    if (i < 10) itemsFound.set(query, 20);
    else if (itemsFound.has(query)) itemsFound.set(query, 20);
    i++;
  }

  for (const [key, value] of itemsFound) {
    const newValue = value - 1;
    if (newValue <= 0) itemsFound.delete(key);
    else itemsFound.set(key, newValue);
  }
}

export async function scanCs(): Promise<DealNotification[]> {
  if (!globals.CS_ITEMS) return [];

  const notifications: DealNotification[] = [];

  const scanResult = await resultAsync(async () => {
    const item = await getCsMarketQuery();
    if (!item) return;

    await sleep(3000);

    const listingResult = await fetchSteamCsMarketListing(item.url);
    if (listingResult.isErr()) {
      console.warn(
        "Failed to fetch CS market listing:",
        listingResult.error.message,
      );
      return;
    }

    const listingData = listingResult.value;
    await processCsMarketListings(item, listingData, notifications);
  }, "CS market scan failed");

  if (scanResult.isErr()) {
    console.error(scanResult.error.message);
  }

  return notifications;
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
  const createResult = await resultAsync(async () => {
    if (!oldQuery.includes("https://steamcommunity.com/market/listings/730/")) {
      return "";
    }

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
            maxPrice,
            maxFloat,
            lastPrice: 0,
          },
        },
      },
    });

    return search;
  }, "Failed to create CS market query");

  if (createResult.isErr()) {
    console.error(createResult.error.message);
    return "";
  }

  return createResult.value;
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
