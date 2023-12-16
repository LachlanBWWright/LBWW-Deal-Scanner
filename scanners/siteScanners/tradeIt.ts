import { Client, TextChannel } from "discord.js";
import axios from "axios";
import TradeItItem from "../../schema/tradeItItem.js";
import { JSONArray } from "puppeteer";

class TradeIt {
  wasFound: boolean;
  client: Client;
  channelId: string;
  roleId: string;
  cursor: any;

  constructor(client: Client, channelId: string, roleId: string) {
    this.wasFound = false;
    this.client = client;
    this.channelId = channelId;
    this.roleId = roleId;
    this.cursor = TradeItItem.find().cursor();

    this.scan = this.scan.bind(this);
    this.skinExists = this.skinExists.bind(this);
  }

  async scan() {
    let itemsArray: JSONArray = [];
    for (let i = 0; i < 20; i++) {
      //Has to make multiple searches due to a size limit.
      await axios
        .get(
          `https://tradeit.gg/api/v2/inventory/data?gameId=730&offset=${
            i * 1000
          }&limit=1000&sortType=(CSGO)+Best+Float&searchValue=&minPrice=0&maxPrice=100000&minFloat=0&maxFloat=1&hideTradeLock=false&fresh=true`,
          {
            headers: {
              "User-Agent":
                "Mozilla/5.0 (platform; rv:geckoversion) Gecko/geckotrail Firefox/firefoxversion",
            },
          }
        )
        .then((res) => {
          itemsArray = [...itemsArray, ...res.data.items];
          if (res.data.items.length < 750) i = 20; //Breaks the loop if it's reached the end of the item list
        })
        .catch((e) => console.error(e));
    }
    try {
      let items = <any>itemsArray;
      let cursor = TradeItItem.find().cursor();
      for (
        let item = await cursor.next();
        item != null;
        item = await cursor.next()
      ) {
        let itemWasFound = false;
        let itemCount = items.length;
        for (let i = 0; i < itemCount; i++) {
          if (
            items[i].price / 100.0 <= item.maxPrice &&
            items[i].name === item.name
          ) {
            let bestFloat = 1;
            if (items[i].floatValue) bestFloat = items[i].floatValue;
            else if (items[i].floatValues) {
              for (let x = 0; x < items[i].floatValues.length; x++) {
                if (items[i].floatValues[x] < bestFloat)
                  bestFloat = items[i].floatValues[x];
              }
            }
            if (bestFloat >= item.minFloat && bestFloat <= item.maxFloat) {
              if (!item.found) {
                //This stops repeated notification messages; the skin must not appear in a search for another message to be sent
                item.found = true;
                item.save((e) => console.error(e));

                this.client.channels
                  .fetch(this.channelId)
                  .then((channel) => <TextChannel>channel)
                  .then((channel) => {
                    if (channel)
                      channel.send(
                        `<@&${this.roleId}> Please know that a ${
                          items[i].name
                        } with a float of ${bestFloat} is available for $${
                          items[i].price / 100.0
                        } USD at: https://tradeit.gg/csgo/trade`
                      );
                  })
                  .catch((e) => console.error(e));
              }
              itemWasFound = true;
            }
          }
        }
        if (!itemWasFound) {
          item.found = false;
          item.save();
        }
      }
    } catch (e) {
      console.error(e);
    }
  }

  async skinExists(name: string) {
    let itemsArray: JSONArray = [];
    let skinFound = false;
    for (let i = 0; i < 20; i++) {
      //Has to make multiple searches due to a size limit.
      await axios
        .get(
          `https://tradeit.gg/api/v2/inventory/data?gameId=730&offset=${
            i * 1000
          }&limit=1000&sortType=(CSGO)+Best+Float&searchValue=&minPrice=0&maxPrice=100000&minFloat=0&maxFloat=1&hideTradeLock=false&fresh=true`,
          {
            headers: {
              "User-Agent":
                "Mozilla/5.0 (platform; rv:geckoversion) Gecko/geckotrail Firefox/firefoxversion",
            },
          }
        )
        .then((res) => {
          itemsArray = [...itemsArray, ...res.data.items];
          if (res.data.items.length < 750) i = 20; //Breaks the loop if it's reached the end of the item list
        })
        .catch((e) => console.error(e));
    }
    try {
      let items = <any>itemsArray;
      for (let i = 0; i < items.length; i++) {
        if (items[i].name == name) {
          skinFound = true;
          break;
        }
      }
    } catch (e) {
      console.error(e);
    }
    return skinFound;
  }
}

export default TradeIt;

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
