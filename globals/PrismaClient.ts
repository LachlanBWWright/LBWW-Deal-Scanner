import { PrismaClient } from "@prisma/client";
export const db = new PrismaClient();

export enum SCANNER {
  CASH_CONVERTERS = 0,
  EBAY = 1,
  GUMTREE = 2,
  SALVOS = 3,
  STEAM_MARKET_QUERY = 4,
  STEAM_MARKET_CS = 5,
  CS_TRADE_BOT = 6,
}

console.log(SCANNER.EBAY);
