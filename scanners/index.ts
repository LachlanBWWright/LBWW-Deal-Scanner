import { scanCSDeals } from "./siteScanners/csDeals.js";
import { scanCSTrade } from "./siteScanners/csTrade.js";
import { scanTradeIt } from "./siteScanners/tradeIt.js";
import { scanLootFarm } from "./siteScanners/lootFarm.js";
import { scanSteamQuery } from "./siteScanners/steamMarket.js";
import { scanCashConverters } from "./siteScanners/cashConverters.js";
import { scanSalvos } from "./siteScanners/salvos.js";
import { scanEbay } from "./siteScanners/ebay.js";
import { scanGumtree } from "./siteScanners/gumtree.js";
import puppeteer from "puppeteer";

export default async function () {
  console.log("Discord client is ready.");

  const browser = await puppeteer.launch({
    headless: true,
    args: ["--no-sandbox"],
  });
  const page = await browser.newPage();

  //For Restricting how often certain scans are peformed to avoid rate-limiting
  let steamScanCnt = 0;
  let csTradeScanCnt = 0;

  //Round-robin scanning
  while (true) {
    console.log("Cycle Start");
    console.time("Cycle Time");

    //Scans performed every iteration
    console.log("Cash Converters");
    await scanCashConverters(page);
    console.log("Ebay");
    await scanEbay(page);
    console.log("Gumtree");
    await scanGumtree(page);
    console.log("Salvos");
    await scanSalvos(page);

    //Scans performed at limited intervals
    if (steamScanCnt >= 55) {
      await scanSteamQuery();
      steamScanCnt = 0;
    }
    if (csTradeScanCnt >= 100) {
      //All these are all at once, only done every 100 cycles
      await scanCSDeals(page);
      await scanCSTrade();
      await scanLootFarm();
      await scanTradeIt();
      csTradeScanCnt = 0;
    }

    steamScanCnt++;
    csTradeScanCnt++;
    console.timeEnd("Cycle Time");
    console.log("Cycle End");
  }
}
