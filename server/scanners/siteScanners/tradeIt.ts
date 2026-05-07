import axios from "axios";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import {
  checkIfNewCsItem,
  CsSite,
  getAllTradeBotItems,
} from "../../functions/csTradeBot.js";
import { fromThrowableAsync } from "../../functions/neverthrowUtils.js";
import type { DealNotification } from "../../deals/types.js";

export async function scanTradeIt(): Promise<DealNotification[]> {
  if (!globals.CS_ITEMS) return [];
  setStatus("Scanning tradeit.gg");

  interface TradeItItem {
    price: number;
    name: string;
    floatValue?: number;
    floatValues?: number[];
    // allow other unknown props from the API
    [k: string]: unknown;
  }

  let itemsArray: TradeItItem[] = [];
  for (let i = 0; i < 20; i++) {
    //Has to make multiple searches due to a size limit.
    const batchResult = await fromThrowableAsync(
      () =>
        axios.get(
          `https://tradeit.gg/api/v2/inventory/data?gameId=730&offset=${
            i * 1000
          }&limit=1000&sortType=(CSGO)+Best+Float&searchValue=&minPrice=0&maxPrice=100000&minFloat=0&maxFloat=1&hideTradeLock=false&fresh=true`,
          {
            headers: {
              "User-Agent":
                "Mozilla/5.0 (platform; rv:geckoversion) Gecko/geckotrail Firefox/firefoxversion",
            },
          },
        ),
      "Failed to fetch tradeit batch",
    );

    if (batchResult.isErr()) {
      continue;
    }

    const res = batchResult.value;
    // Validate items structure
    const items = res.data.items;
    if (Array.isArray(items)) {
      const validItems = items.filter(
        (item: unknown): item is TradeItItem =>
          typeof item === "object" &&
          item !== null &&
          "price" in item &&
          "name" in item,
      );

      itemsArray = [...itemsArray, ...validItems];
      if (items.length < 750) break;
    }
  }

  const notifications: DealNotification[] = [];

  const scanResult = await fromThrowableAsync(async () => {
    const foundItems = itemsArray;
    const searchItems = await getAllTradeBotItems();

    for (const searchItem of searchItems) {
      for (const foundItem of foundItems) {
        if (
          foundItem.price / 100.0 <= searchItem.maxPrice &&
          foundItem.name === searchItem.name
        ) {
          let bestFloat = 1;
          if (foundItem.floatValue) bestFloat = foundItem.floatValue;
          else if (foundItem.floatValues) {
            for (const floatVal of foundItem.floatValues) {
              if (floatVal < bestFloat) bestFloat = floatVal;
            }
          }

          if (
            bestFloat >= searchItem.minFloat &&
            bestFloat <= searchItem.maxFloat
          ) {
            if (
              await checkIfNewCsItem(
                searchItem.name,
                bestFloat,
                CsSite.LOOT_FARM,
              )
            ) {
              notifications.push({
                kind: "deal",
                source: "tradeIt",
                title: `a ${foundItem.name} with a float of ${bestFloat} is available for $${foundItem.price / 100.0} USD at: https://tradeit.gg/csgo/trade`,
                url: "https://tradeit.gg/csgo/trade",
                price: foundItem.price / 100.0,
                query: {
                  type: "csTradeBot",
                  id: searchItem.name,
                },
              });
            }
          }
        }
      }
    }
  }, "TradeIt scan failed");

  if (scanResult.isErr()) {
    console.error(scanResult.error.message);
  }

  return notifications;
}

/* {
    "id":"23712770392",
    "assetId":"23712770392",
    "classId":"2980086889",
    "steamId":"76561198899818391",
    "assetLength":1,
    "price":20950,
    "botIndex":"75",
    "floatValue":0.00337359,
    "paintIndex":44,
    "metaMappings":{
    "rarity":4,
    "type":6
    },
    "gameId":"CSGO",
    "imgURL":"https://old.tradeit.gg/static/img/items/316911.png",
    "name":"★ Navaja Knife | Case Hardened (Factory New)",
    "phase":null,
    "score":225350,
    "wantedStock":2,
    "currentStock":1,
    "steamAppId":730,
    "steamContextId":2,
    "steamInspectLink":"steam://rungame/730/76561202255233023/+csgo_econ_action_preview%20S76561198899818391A23712770392D7405083392701897930",
    "steamMarketLink":"https://steamcommunity.com/market/listings/730/★ Navaja Knife | Case Hardened (Factory New)",
    "steamInventoryLink":"https://steamcommunity.com/profiles/76561198899818391/inventory/#730_2980086889_23712770392",
    "steamTags":[
    "Knife",
    "Navaja Knife",
    "",
    "★",
    "Covert",
    "Factory New",
    "",
    "",
    "",
    "",
    "",
    "",
    ""
    ],
    "stickers":null,
    "createdAt":"2022-05-04T15:45:10.598Z",
    "tradedAt":"2022-05-04T15:45:10.598Z",
    "groupId":316911,
    "_id":"23712770392"
    } */
