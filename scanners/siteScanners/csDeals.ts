import { Client, TextChannel } from "discord.js";
import puppeteer, { HTTPResponse } from "puppeteer";
import CsDealsItem from "../../schema/csDealsItem.js";
import globals from "../../globals/Globals.js";
import client from "../../globals/DiscordJSClient.js";

export async function scanCSDeals(page: puppeteer.Page) {
  if (!globals.CS_ITEMS || !globals.CS_CHANNEL_ID || !globals.CS_ROLE_ID)
    return;

  try {
    await page.setDefaultNavigationTimeout(0); //TODO: Consider removing this
    page.goto("https://cs.deals/trade-skins");

    //New eventlistener replacement
    let foundResponse;
    await page.waitForResponse((response) => {
      if (response.url().endsWith("botsinventory?appid=0")) {
        foundResponse = response;
        return true;
      } else return false;
    });

    if (foundResponse != undefined) {
      foundResponse = <HTTPResponse>foundResponse;
      let items = await foundResponse.json();
      let csgoItemCount = items.response.items[730].length;
      items = items.response.items[730];
      let cursor = CsDealsItem.find().cursor(); //Iterates through every DB item
      for (
        let item = await cursor.next();
        item != null;
        item = await cursor.next()
      ) {
        let itemWasFound = false;
        for (let i = 0; i < csgoItemCount; i++) {
          //Iterates through every item on the website
          //Checks if a match is found, and sends a message if it is | .c = Name, .d1 = Float, .price = Price
          if (
            items[i].c === item.name &&
            items[i].d1 < item.maxFloat &&
            items[i].d1 > item.minFloat &&
            items[i].i <= item.maxPrice
          ) {
            if (!item.found) {
              //This stops repeated notification messages; the skin must not appear in a search for another message to be sent
              item.found = true;
              item.save((e) => console.error(e));

              client.channels
                .fetch(globals.CS_CHANNEL_ID)
                .then((channel) => <TextChannel>channel)
                .then((channel) => {
                  if (channel)
                    channel.send(
                      `<@&${globals.CS_ROLE_ID}> Please know that a ${items[i].c} with a float of ${items[i].d1} is available for $${items[i].i} USD at: https://cs.deals/trade-skins`
                    );
                })
                .catch((e) => console.error(e));
            }
            itemWasFound = true;
          }
        }
        if (!itemWasFound) {
          item.found = false;
          item.save();
        }
      }
    }
  } catch (e) {
    console.error(e);
  }
}

class CsDeals {
  client: Client;
  channelId: string;
  roleId: string;

  constructor(client: Client, channelId: string, roleId: string) {
    this.client = client;
    this.channelId = channelId;
    this.roleId = roleId;

    this.skinExists = this.skinExists.bind(this);
  }

  async skinExists(name: string) {
    const browser = await puppeteer.launch();
    try {
      const page = await browser.newPage();
      await page.goto("https://cs.deals/trade-skins");

      let skinWasFound = false;
      let foundResponse;

      await page.waitForResponse((response) => {
        if (response.url().endsWith("botsinventory?appid=0")) {
          foundResponse = response;
          return true;
        } else {
          return false;
        }
      });

      if (foundResponse != undefined) {
        foundResponse = <HTTPResponse>foundResponse;
        let items = await foundResponse.json();
        let csgoItemCount = items.response.items[730].length;
        items = items.response.items[730];

        for (let i = 0; i < csgoItemCount; i++) {
          //Iterates through every item on the website
          //Checks if a match is found, and sends a message if it is | .c = Name
          if (items[i].c === name) {
            skinWasFound = true;
            break;
          }
        }
      } else {
        await browser.close();
        return false;
      }
      await browser.close();
      //Returns true if a skin with the matching name was found
      if (skinWasFound) return true;
      else return false;
    } catch (e) {
      console.error(e);
    }
  }
}

export default CsDeals;

/* 
JSON Example
{
  a: '68',
  c: 'Souvenir AUG | Radiation Hazard (Factory New)',
  e: 4,
  g: '3023755629',
  h: '188530139',
  i: 5.82,                                                                                  The actual price
  k: '5e98d9',
  q: 'OTg4ODg=',
  x: 5.43,                                                                                  Some sort of 'base price', lower than actual
  f: '22590511304',
  t: 8499961,
  f1: [
    [
      'Fnatic (Gold) | Cologne 2016',
      'cologne2016/fntc_gold.e622cb6a1885e75f2a1d068efdabba19a6e87d5a.png',
      '0.00000000000000000',
      '0'
    ],
    [
      'ESL (Gold) | Cologne 2016',
      'cologne2016/esl_gold.f2b63efe44b0a777411448aaf9604014ee414e02.png',
      '1.00000000000000000',
      '2'
    ]
  ],
  b1: '5621234145811519616',
  o: 'Rifle',
  h1: 'MTY4MzE4NQ==',
  c1: 'FN',
  d1: '0.05834920331836',                                                                Float Value
  e1: 160
}
*/
