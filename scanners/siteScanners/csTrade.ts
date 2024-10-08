import axios from "axios";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import sendToChannel from "../../functions/sendToChannel.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";
import {
  checkIfNewCsItem,
  CsSite,
  getAllTradeBotItems,
} from "../../functions/csTradeBot.js";

export async function scanCSTrade() {
  if (!globals.CS_ITEMS || !globals.CS_CHANNEL_ID || !globals.CS_ROLE_ID)
    return;
  setStatus("Scanning CS.Trade");

  const res = await axios.get(
    "https://cdn.cs.trade:8443/api/getInventory?order_by=price_desc&bot=all&_=1651756783463",
  );

  const foundItems = await getCsTradeItems();
  const searchItems = await getAllTradeBotItems();

  for (const searchItem of searchItems) {
    for (const foundItem of foundItems) {
      if (
        foundItem.price <= searchItem.maxPrice &&
        foundItem.wear >= searchItem.minFloat &&
        foundItem.wear <= searchItem.maxFloat &&
        foundItem.market_hash_name === searchItem.name
      ) {
        if (
          await checkIfNewCsItem(foundItem.c, foundItem.d1, CsSite.CS_TRADE)
        ) {
          sendToChannel(
            globals.CS_CHANNEL_ID,
            `<@&${globals.CS_ROLE_ID}> ${getNotificationPrelude()} a ${
              foundItem.market_hash_name
            } with a float of ${foundItem.wear} is available for $${
              foundItem.price
            } USD at: https://cs.trade/`,
          );
        }
      }
    }
  }
}

export async function getCsTradeItems() {
  const res = await axios.get(
    "https://cdn.cs.trade:8443/api/getInventory?order_by=price_desc&bot=all&_=1651756783463",
  );

  return res.data.inventory.filter(
    (item: { app_id: number }) => item.app_id == 730,
  );
}

/* {
    id: '25477418560_730',
    app_id: '730',
    market_hash_name: '★ Bayonet | Slaughter (Minimal Wear)',
    price: 506.04,     //PRICE
    icon: 'https://steamcommunity-a.akamaihd.net/economy/image/-9a81dlWLwJ2UUGcVs_nsVtzdOEdtWwKGZZLQHTxDZ7I56KU0Zwwo4NUX4oFJZEHLbXH5ApeO4YmlhxYQknCRvCo04DEVlxkKgpotLu8JAllx8zJfAJY6d6klb-HnvD8J_WDxDgFuJMl2b-Tp9yhjQzjrhJpMDzwco-cdVJtZ1HRrlm6xbrmhJC0ot2XnobxE0h8/330x192',
    status: 'tradable',
    reservable: false,
    status_description: 'Unavailable',
    type: 'CSGO_Type_Knife',
    inspect_link: 'steam://rungame/730/76561202255233023/+csgo_econ_action_preview%20S76561198310607331A25477418560D5639753587697763588',
    stattrak: false,
    souvenir: null,
    skin: true,
    rare: true,
    stattrakknife: false,
    wear: '0.08060055',
    wear_to_display: '0.08060055',
    name_to_display: '★ Bayonet | Slaughter',
    wear_name: 'MW',
    stickers: null,
    paint_index: null,
    doppler: false,
    float_bonus: 0,
    tradable_bool: false,
    marketable_bool: true,
    name_color: '#8650AC',
    dota2_rarity: null,
    dota2_type: null,
    dota2_hero: null,
    h1z1_slot: null,
    h1z1_rarity: null,
    rust_type: null,
    bot: '6',
    bot_id: '76561198310607331',
    tradable_from: {
      date: '2022-04-26 09:00:00.000000',
      timezone_type: 3,
      timezone: 'Europe/Warsaw'
    }
  }, */
