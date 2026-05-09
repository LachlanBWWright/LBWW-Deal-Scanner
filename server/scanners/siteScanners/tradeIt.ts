import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import {
  checkIfNewCsItem,
  CsSite,
  getAllTradeBotItems,
} from "../../functions/csTradeBot.js";
import { resultAsync } from "../../functions/neverthrowUtils.js";
import { fetchTradeItAllItems } from "../../functions/tradeItFetcher.js";
import type { DealNotification } from "../../deals/types.js";
import type { TradeItItem } from "../../functions/apiValidators.js";

function getBestFloat(foundItem: TradeItItem): number {
  if (foundItem.floatValue) return foundItem.floatValue;

  if (foundItem.floatValues && foundItem.floatValues.length > 0) {
    return Math.min(...foundItem.floatValues);
  }

  return 1;
}

async function checkTradeItMatch(
  searchItem: Awaited<ReturnType<typeof getAllTradeBotItems>>[0],
  foundItem: TradeItItem,
  notifications: DealNotification[],
) {
  if (
    foundItem.price / 100.0 > searchItem.maxPrice ||
    foundItem.name !== searchItem.name
  ) {
    return;
  }

  const bestFloat = getBestFloat(foundItem);
  if (bestFloat < searchItem.minFloat || bestFloat > searchItem.maxFloat) {
    return;
  }

  const isNew = await checkIfNewCsItem(
    searchItem.name,
    bestFloat,
    CsSite.TRADEIT_GG,
  );
  if (!isNew) return;

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

export async function scanTradeIt(): Promise<DealNotification[]> {
  if (!globals.CS_ITEMS) return [];
  setStatus("Scanning tradeit.gg");

  const itemsArray = await fetchTradeItAllItems();

  const notifications: DealNotification[] = [];

  const scanResult = await resultAsync(async () => {
    const searchItems = await getAllTradeBotItems();

    for (const searchItem of searchItems) {
      for (const foundItem of itemsArray) {
        await checkTradeItMatch(searchItem, foundItem, notifications);
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
