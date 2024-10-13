import axios from "axios";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import sendToChannel from "../../functions/sendToChannel.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";
import { db } from "../../globals/PrismaClient.js";
import {
  checkIfNewCsItem,
  CsSite,
  getAllTradeBotItems,
} from "../../functions/csTradeBot.js";

export async function scanLootFarm() {
  if (!globals.CS_ITEMS || !globals.CS_CHANNEL_ID || !globals.CS_ROLE_ID)
    return;
  setStatus("Scanning loot.farm");

  let items = await getLootFarmItems();

  const searchItems = await getAllTradeBotItems();

  for (const searchItem of searchItems) {
    for (let skinType in items) {
      if (
        searchItem.name.includes(items[skinType].n) &&
        items[skinType].p / 100 <= searchItem.maxPrice
      ) {
        for (let botNumber in items[skinType].u) {
          const item = items[skinType].u[botNumber][0];
          const itemFloat = parseInt(item.f) / 100000;
          if (
            itemFloat >= searchItem.minFloat &&
            itemFloat <= searchItem.maxFloat &&
            /* If searchItem is StatTrak, check that the found item is also StatTrak */
            (!searchItem.name.includes("StatTrak") || item.st != undefined)
          ) {
            if (
              await checkIfNewCsItem(searchItem.name, item.f, CsSite.LOOT_FARM)
            ) {
              sendToChannel(
                globals.CS_CHANNEL_ID,
                `<@&${globals.CS_ROLE_ID}> ${getNotificationPrelude()} a ${
                  items[skinType].n
                } with a float of ${itemFloat} is available for $${
                  items[skinType].p / 100
                } USD at: https://loot.farm/`,
              );
            }
          }
        }
      }
    }
  }
}

export async function getLootFarmItems() {
  const res = await axios.get("https://loot.farm/botsInventory_730.json");
  return res.data.result;
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
