import { describe, expect, test } from "vitest";

import { fetchCsTradeItems } from "./csTraceFetcher.js";
import { fetchLootFarmItems } from "./lootFarmFetcher.js";
import {
  fetchSteamCsMarketListing,
  fetchSteamItemInfo,
  fetchSteamMarketResults,
} from "./steamMarketFetcher.js";
import { fetchTradeItItemsBatch } from "./tradeItFetcher.js";

const runLiveApiTests = process.env.RUN_LIVE_API_TESTS === "true";

describe.skipIf(!runLiveApiTests)("live API validation", () => {
  test("validates loot.farm inventory", async () => {
    const result = await fetchLootFarmItems();

    expect(result.isOk()).toBe(true);
    if (result.isErr()) return;

    expect(Object.keys(result.value.result).length).toBeGreaterThan(0);
  }, 60000);

  test("validates CS.Trade inventory", async () => {
    const result = await fetchCsTradeItems();

    expect(result.isOk()).toBe(true);
    if (result.isErr()) return;

    expect(result.value.length).toBeGreaterThan(0);
    expect(result.value[0]?.app_id).toBe(730);
  }, 60000);

  test("validates TradeIt inventory batch", async () => {
    const result = await fetchTradeItItemsBatch(0);

    expect(result.isOk()).toBe(true);
    if (result.isErr()) return;

    expect(result.value.length).toBeGreaterThan(0);
    expect(typeof result.value[0]?.name).toBe("string");
  }, 60000);

  test("validates Steam market search results", async () => {
    const result = await fetchSteamMarketResults(
      "https://steamcommunity.com/market/search/render/?query=AK-47&start=0&count=2&norender=1",
    );

    expect(result.isOk()).toBe(true);
    if (result.isErr()) return;

    expect(result.value.length).toBeGreaterThan(0);
    expect(typeof result.value[0]?.sell_price).toBe("number");
  }, 60000);

  test("validates Steam listing and CSGOFloat item info", async () => {
    const listingResult = await fetchSteamCsMarketListing(
      "https://steamcommunity.com/market/listings/730/AK-47%20%7C%20Redline%20%28Field-Tested%29/render/?query=&start=0&count=2&country=AU&language=english&currency=1",
    );

    expect(listingResult.isOk()).toBe(true);
    if (listingResult.isErr()) return;

    expect(Object.keys(listingResult.value.listinginfo).length).toBeGreaterThanOrEqual(0);

    const tradeItResult = await fetchTradeItItemsBatch(0);
    expect(tradeItResult.isOk()).toBe(true);
    if (tradeItResult.isErr()) return;

    const inspectLink = tradeItResult.value.find((item) => item.steamInspectLink)?.steamInspectLink;
    expect(inspectLink).toBeTruthy();
    if (!inspectLink) return;

    const itemInfoUrl = `https://api.csgofloat.com/?url=${encodeURIComponent(inspectLink)}`;

    const itemInfoResult = await fetchSteamItemInfo(itemInfoUrl);
    if (itemInfoResult.isErr()) {
      expect(itemInfoResult.error.message).toContain("429 Too Many Requests");
      return;
    }

    expect(typeof itemInfoResult.value.iteminfo.floatvalue).toBe("number");
    expect(typeof itemInfoResult.value.iteminfo.full_item_name).toBe("string");
  }, 60000);
});