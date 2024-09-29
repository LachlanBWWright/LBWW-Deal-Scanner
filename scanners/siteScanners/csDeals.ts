import puppeteer, { HTTPResponse } from "puppeteer";
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

export async function scanCSDeals(page: puppeteer.Page) {
  if (!globals.CS_ITEMS || !globals.CS_CHANNEL_ID || !globals.CS_ROLE_ID)
    return;
  setStatus("Scanning CS Deals");

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

  if (!foundResponse) return;
  foundResponse = <HTTPResponse>foundResponse;

  const foundItems = (await foundResponse.json()).response.items[730];
  const searchItems = await getAllTradeBotItems();

  for (const searchItem of searchItems) {
    for (const foundItem of foundItems) {
      //Iterates through every item on the website
      //Checks if a match is found, and sends a message if it is | .c = Name, .d1 = Float, .price = Price
      if (
        foundItem.c === searchItem.name &&
        foundItem.d1 < searchItem.maxFloat &&
        foundItem.d1 > searchItem.minFloat &&
        foundItem.i <= searchItem.maxPrice
      ) {
        if (await checkIfNewCsItem(foundItem.c, foundItem.d1, CsSite.CS_DEALS))
          sendToChannel(
            globals.CS_CHANNEL_ID,
            `<@&${globals.CS_ROLE_ID}> ${getNotificationPrelude()} a ${
              foundItem.c
            } with a float of ${foundItem.d1} is available for $${
              foundItem.i
            } USD at: https://cs.deals/trade-skins`,
          );
      }
    }
  }
}

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
