import { scanCSTrade } from "./siteScanners/csTrade.js";
import { scanTradeIt } from "./siteScanners/tradeIt.js";
import { scanLootFarm } from "./siteScanners/lootFarm.js";
import { scanSteamQuery } from "./siteScanners/steamMarket.js";
import { scanCashConverters } from "./siteScanners/cashConverters.js";
import { scanSalvos } from "./siteScanners/salvos.js";
import { scanEbay } from "./siteScanners/ebay.js";
import { scanGumtree } from "./siteScanners/gumtree.js";
import handleError from "../functions/handleError.js";
import puppeteer from "puppeteer";
import { memoryUsage } from "node:process";

export default async function () {
  console.log("Discord client is ready.");

  const browser = await puppeteer.launch({
    headless: "shell",
    args: ["--no-sandbox"],
  });
  let page = await browser.newPage();

  //For Restricting how often certain scans are peformed to avoid rate-limiting
  let steamScanCnt = 0;
  let csTradeScanCnt = 0;

  //Round-robin scanning
  while (true) {
    console.log(memoryUsage());
    const throttler = new Promise((r) => setTimeout(r, 2000));
    console.time("Cycle Time (2000ms minimum)");

    //Scans performed every iteration
    await scanCashConverters(page).catch((err) =>
      handleError(err, "Cash Converters"),
    );
    await scanEbay(page).catch((err) => handleError(err, "Ebay"));
    await scanGumtree(page).catch((err) => handleError(err, "Gumtree"));
    await scanSalvos(page).catch((err) => handleError(err, "Salvos"));

    //Scans performed at limited intervals
    if (steamScanCnt >= 55) {
      await scanSteamQuery();
      steamScanCnt = 0;
    }
    if (csTradeScanCnt >= 100) {
      //All these are all at once, only done every 100 cycles
      await scanCSTrade().catch((err) => handleError(err, "CS Trade"));
      await scanLootFarm().catch((err) => handleError(err, "Loot Farm"));
      await scanTradeIt().catch((err) => handleError(err, "Trade It"));
      csTradeScanCnt = 0;
    }

    steamScanCnt++;
    csTradeScanCnt++;

    await page.close();
    page = await browser.newPage();

    //Throttle
    await throttler;
    console.timeEnd("Cycle Time (2000ms minimum)");
  }
}
