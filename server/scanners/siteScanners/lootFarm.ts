import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import {
  checkIfNewCsItem,
  CsSite,
  getAllTradeBotItems,
} from "../../functions/csTradeBot.js";
import { fetchLootFarmItems } from "../../functions/lootFarmFetcher.js";
import type { DealNotification } from "../../deals/types.js";
import type { LootFarmSkin } from "../../functions/apiValidators.js";

async function checkLootFarmItem(
  searchItem: Awaited<ReturnType<typeof getAllTradeBotItems>>[0],
  skin: LootFarmSkin,
  skinPriceCents: number,
  botNumber: string,
  notifications: DealNotification[]
) {
  const item = skin.u[botNumber][0];
  if (!item?.f) return;

  const itemFloat = parseInt(item.f) / 100000;

  const meetsFloatCriteria =
    itemFloat >= searchItem.minFloat && itemFloat <= searchItem.maxFloat;
  const meetsStatTrakCriteria =
    !searchItem.name.includes("StatTrak") || item.st != null;

  if (!meetsFloatCriteria || !meetsStatTrakCriteria) return;

  const isNew = await checkIfNewCsItem(
    searchItem.name,
    itemFloat,
    CsSite.LOOT_FARM
  );
  if (!isNew) return;

  notifications.push({
    kind: "deal",
    source: "lootFarm",
    title: `a ${skin.n} with a float of ${itemFloat} is available for $${skinPriceCents / 100} USD at: https://loot.farm/`,
    url: "https://loot.farm/",
    price: skinPriceCents / 100,
    query: {
      type: "csTradeBot",
      id: searchItem.name,
    },
  });
}

export async function scanLootFarm(): Promise<DealNotification[]> {
  if (!globals.CS_ITEMS) return [];
  setStatus("Scanning loot.farm");

  const itemsResult = await fetchLootFarmItems();
  if (itemsResult.isErr()) {
    console.error("Failed to fetch loot.farm items:", itemsResult.error.message);
    return [];
  }

  const items = itemsResult.value.result;
  const searchItems = await getAllTradeBotItems();
  const notifications: DealNotification[] = [];

  for (const searchItem of searchItems) {
    for (const skinType in items) {
      const skin = items[skinType];
      if (typeof skin.p !== "number") {
        continue;
      }

      const skinPriceCents = skin.p;

      if (
        !searchItem.name.includes(skin.n)
        || skinPriceCents / 100 > searchItem.maxPrice
      ) {
        continue;
      }

      for (const botNumber in skin.u) {
        await checkLootFarmItem(
          searchItem,
          skin,
          skinPriceCents,
          botNumber,
          notifications,
        );
      }
    }
  }

  return notifications;
}

/* "27623861":{
    "n":"AUG | Sweeper",
    "cl":"4842901053",
    "g":1,
    "t":{
    "t":"R",
    "r":"WC"
    },
    "e":"FN",
    "u":{
    "8":[
    {
    "id":"25709717542",
    "f":"5642:100",
    "l":"D13836866265007789130",
    "tr":0,
    "td":140
    }
    ],
    "12":[
    {
    "id":"25651032409",
    "f":"3373:348",
    "l":"D4958081568550984718",
    "tr":1
    }
    ],
    "15":[
    {
    "id":"25634419586",
    "f":"6921:601",
    "l":"D4958081568550984718",
    "tr":1
    }
    ],
    "40":[
    {
    "id":"25480974285",
    "f":"4934:800",
    "l":"D4958081568550984718",
    "tr":1
    }
    ],
    "43":[
    {
    "id":"25649926256",
    "f":"6088:596",
    "l":"D4958081568550984718",
    "tr":1
    },
    {
    "id":"25575668029",
    "f":"6418:410",
    "l":"D4958081568550984718",
    "tr":1
    }
    ],
    "44":[
    {
    "id":"25711228075",
    "f":"4996:382",
    "l":"D13836866265007789130",
    "tr":0,
    "td":140
    }
    ],
    "45":[
    {
    "id":"25510566360",
    "f":"6591:59",
    "l":"D4958081568550984718",
    "tr":1
    }
    ],
    "47":[
    {
    "id":"25476569538",
    "f":"5168:457",
    "l":"D4958081568550984718",
    "tr":1
    }
    ],
    "49":[
    {
    "id":"25555997184",
    "f":"6061:104",
    "l":"D4958081568550984718",
    "tr":1
    }
    ],
    "57":[
    {
    "id":"25498611050",
    "f":"6169:899",
    "l":"D4958081568550984718",
    "tr":1
    }
    ],
    "58":[
    {
    "id":"25573861600",
    "f":"5259:738",
    "l":"D4958081568550984718",
    "tr":1
    }
    ],
    "61":[
    {
    "id":"25594387424",
    "f":"6153:579",
    "l":"D4958081568550984718",
    "tr":1
    }
    ],
    "62":[
    {
    "id":"25465977303",
    "f":"4819:387",
    "l":"D4958081568550984718",
    "tr":1
    }
    ]
    },
    "pg":95,
    "p":5
    }, */
