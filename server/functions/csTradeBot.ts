import { db, SCANNER } from "../globals/PrismaClient.js";

/* Returns all trade bot items to be found */
export async function getAllTradeBotItems() {
  return await db.csTradeBot.findMany();
}

export enum CsSite {
  CS_TRADE = "CS_TRADE",
  CS_DEALS = "CS_DEALS",
  LOOT_FARM = "LOOT_FARM",
  TRADEIT_GG = "TRADEIT_GG",
}

/* 
Checks an item. If it doesn't, add it to the database, and return true.
If it already exists, extend the TTL. Avoid adding items that don't match the user query
*/
export async function checkIfNewCsItem(
  itemName: string,
  float: number,
  csSite: CsSite,
) {
  const item = await db.ttlItem.findFirst({
    where: {
      itemId: `${itemName}_${float}_${csSite}`,
      scanner: SCANNER.CS_TRADE_BOT,
    },
  });
  if (item) {
    await db.ttlItem.update({
      where: {
        itemId: `${itemName}_${float}_${csSite}`,
        scanner: SCANNER.CS_TRADE_BOT,
      },
      data: {
        lastUpdated: new Date(),
      },
    });
    return false;
  }

  return true;
}
