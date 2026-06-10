import { TradeItResponseSchema, type TradeItItem } from "./apiValidators.js";
import { fetchAndValidateJson } from "./fetchJson.js";

const USER_AGENT =
  "Mozilla/5.0 (platform; rv:geckoversion) Gecko/geckotrail Firefox/firefoxversion";

export async function fetchTradeItItemsBatch(batchIndex: number) {
  const offset = batchIndex * 1000;
  const url = new URL("https://tradeit.gg/api/v2/inventory/data");

  url.searchParams.set("gameId", "730");
  url.searchParams.set("offset", offset.toString());
  url.searchParams.set("limit", "1000");
  url.searchParams.set("sortType", "(CSGO)+Best+Float");
  url.searchParams.set("searchValue", "");
  url.searchParams.set("minPrice", "0");
  url.searchParams.set("maxPrice", "100000");
  url.searchParams.set("minFloat", "0");
  url.searchParams.set("maxFloat", "1");
  url.searchParams.set("hideTradeLock", "false");
  url.searchParams.set("fresh", "true");

  return fetchAndValidateJson(
    url.toString(),
    TradeItResponseSchema,
    {
      headers: {
        "User-Agent": USER_AGENT,
      },
    },
    `Failed to fetch and validate TradeIt batch ${batchIndex}`,
  ).map((data) => data.items);
}

export async function fetchTradeItAllItems() {
  const allItems: TradeItItem[] = [];

  for (let i = 0; i < 20; i++) {
    const batchResult = await fetchTradeItItemsBatch(i);

    if (batchResult.isErr()) {
      console.error(`TradeIt batch ${i} failed:`, batchResult.error.message);
      continue;
    }

    const items = batchResult.value;
    allItems.push(...items);

    // If we got fewer than 750 items, we've reached the end
    if (items.length < 750) {
      break;
    }
  }

  return allItems;
}
