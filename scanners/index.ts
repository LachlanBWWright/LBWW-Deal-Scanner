import CsDeals from "./siteScanners/csDeals";
import CsTrade from "./siteScanners/csTrade";
import TradeIt from "./siteScanners/tradeIt";
import LootFarm from "./siteScanners/lootFarm";
import SteamMarket from "./siteScanners/steamMarket";
import CashConverters from "./siteScanners/cashConverters";
import Salvos from "./siteScanners/salvos";
import Ebay from "./siteScanners/ebay";
import Gumtree from "./siteScanners/gumtree";

import { Client } from "discord.js";
import puppeteer from "puppeteer";
import Dotenv from "dotenv";
Dotenv.config();

export default async function (client: Client) {
  console.log("Discord client is ready.");

  const csDeals = new CsDeals(
    client,
    `${process.env.CS_CHANNEL_ID}`,
    `${process.env.CS_ROLE_ID}`
  );
  const csTrade = new CsTrade(
    client,
    `${process.env.CS_CHANNEL_ID}`,
    `${process.env.CS_ROLE_ID}`
  );
  const tradeIt = new TradeIt(
    client,
    `${process.env.CS_CHANNEL_ID}`,
    `${process.env.CS_ROLE_ID}`
  );
  const lootFarm = new LootFarm(
    client,
    `${process.env.CS_CHANNEL_ID}`,
    `${process.env.CS_ROLE_ID}`
  );
  const steamMarket = new SteamMarket(
    client,
    `${process.env.STEAM_QUERY_CHANNEL_ID}`,
    `${process.env.STEAM_QUERY_ROLE_ID}`,
    `${process.env.CS_MARKET_CHANNEL_ID}`,
    `${process.env.CS_MARKET_ROLE_ID}`
  );
  const cashConverters = new CashConverters(
    client,
    `${process.env.CASH_CONVERTERS_CHANNEL_ID}`,
    `${process.env.CASH_CONVERTERS_ROLE_ID}`
  );
  const salvos = new Salvos(
    client,
    `${process.env.SALVOS_CHANNEL_ID}`,
    `${process.env.SALVOS_ROLE_ID}`
  );
  const ebay = new Ebay(
    client,
    `${process.env.EBAY_CHANNEL_ID}`,
    `${process.env.EBAY_ROLE_ID}`
  );
  const gumtree = new Gumtree(
    client,
    `${process.env.GUMTREE_CHANNEL_ID}`,
    `${process.env.GUMTREE_ROLE_ID}`
  );

  const scanInfrequently = async () => {
    const browser = await puppeteer.launch({
      headless: true,
      args: ["--no-sandbox"],
    });
    const page = await browser.newPage();

    let scanCashConverters =
      process.env.CASH_CONVERTERS === "true" &&
      process.env.CASH_CONVERTERS_CHANNEL_ID &&
      process.env.CASH_CONVERTERS_ROLE_ID;
    let scanGumtree =
      process.env.GUMTREE === "true" &&
      process.env.GUMTREE_CHANNEL_ID &&
      process.env.GUMTREE_ROLE_ID;
    let scanSalvos =
      process.env.SALVOS === "true" &&
      process.env.SALVOS_CHANNEL_ID &&
      process.env.SALVOS_ROLE_ID;
    let scanEbay =
      process.env.EBAY === "true" &&
      process.env.EBAY_CHANNEL_ID &&
      process.env.EBAY_ROLE_ID;
    let scanSteam =
      process.env.STEAM_QUERY === "true" &&
      process.env.STEAM_QUERY_CHANNEL_ID &&
      process.env.STEAM_QUERY_ROLE_ID &&
      process.env.CS_MARKET_CHANNEL_ID &&
      process.env.CS_MARKET_ROLE_ID;
    let scanCs =
      process.env.CS_ITEMS === "true" &&
      process.env.CS_CHANNEL_ID &&
      process.env.CS_ROLE_ID;

    //For Restricting how often certain scans are peformed to avoid rate-limiting
    let steamScanCnt = 0;
    let csTradeScanCnt = 0;

    //Round-robin scanning
    while (true) {
      console.time("Cycle Time");

      //Scans performed every iteration
      if (scanCashConverters) await cashConverters.scan(page);
      if (scanEbay) await ebay.scan(page);
      if (scanGumtree) await gumtree.scan(page);
      if (scanSalvos) await salvos.scan(page);

      //Scans performed at limited intervals
      if (scanSteam && steamScanCnt >= 55) {
        await steamMarket.scanQuery(); //Non-puppeteer
        steamScanCnt = 0;
      }
      if (scanCs && csTradeScanCnt >= 100) {
        //All these are all at once, only done every 100 cycles
        await csDeals.scan(page);
        await csTrade.scan(); //Non-puppeteer
        await lootFarm.scan(); //Non-puppeteer
        await tradeIt.scan(); //Non-puppeteer
        csTradeScanCnt = 0;
      }

      steamScanCnt++;
      csTradeScanCnt++;
      console.timeEnd("Cycle Time");
    }
  };
  scanInfrequently();
}
