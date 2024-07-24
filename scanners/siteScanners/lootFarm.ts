import axios from "axios";
import LootFarmItem from "../../schema/lootFarmItem.js";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import sendToChannel from "../../functions/sendToChannel.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";

export async function scanLootFarm() {
  if (!globals.CS_ITEMS || !globals.CS_CHANNEL_ID || !globals.CS_ROLE_ID)
    return;
  setStatus("Scanning loot.farm");

  const res = await axios.get("https://loot.farm/botsInventory_730.json");
  let items: any = res.data.result;
  let cursor = LootFarmItem.find().cursor();
  for (
    let item = await cursor.next();
    item != null;
    item = await cursor.next()
  ) {
    let itemWasFound = false;

    for (let skinType in items) {
      if (
        item.name.includes(items[skinType].n) &&
        items[skinType].p / 100 <= item.maxPrice
      ) {
        for (let instance in items[skinType].u) {
          if (
            parseInt(items[skinType].u[instance][0].f) / 100000 >=
              item.minFloat &&
            parseInt(items[skinType].u[instance][0].f) / 100000 <=
              item.maxFloat &&
            (!item.name.includes("StatTrak") ||
              items[skinType].u[instance][0].st != undefined)
          ) {
            if (!item.found) {
              //This stops repeated notification messages; the skin must not appear in a search for another message to be sent
              item.found = true;
              item.save((e) => console.error(e));

              sendToChannel(
                globals.CS_CHANNEL_ID,
                `<@&${globals.CS_ROLE_ID}> ${getNotificationPrelude()} a ${
                  items[skinType].n
                } with a float of ${
                  parseInt(items[skinType].u[instance][0].f) / 100000
                } is available for $${
                  items[skinType].p / 100
                } USD at: https://loot.farm/`,
              );
            }
            itemWasFound = true;
          }
        }
      }
    }
    if (!itemWasFound) {
      item.found = false;
      item.save();
    }
  }
}

export async function lootFarmSkinExists(name: string) {
  let skinFound = false;
  await axios
    .get("https://loot.farm/botsInventory_730.json")
    .then(async (res) => {
      let items = res.data.result;

      for (let skinType in items) {
        if (name.includes(items[skinType].n)) {
          skinFound = true;
        }
      }
    })
    .catch((e) => console.error(e));
  return skinFound;
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
