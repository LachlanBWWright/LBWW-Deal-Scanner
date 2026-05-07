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
import { fromThrowableAsync } from "../functions/neverthrowUtils.js";
import type { DealNotification } from "../deals/types.js";

export default async function () {
  console.log("Discord client is ready.");

  const browser = await puppeteer.launch({
    args: ["--no-sandbox"],
    headless: true,
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
    await handleScan(() => scanCashConverters(page), "Cash Converters");
    await handleScan(() => scanEbay(page), "Ebay");
    await handleScan(() => scanGumtree(page), "Gumtree");
    await handleScan(() => scanSalvos(page), "Salvos");

    //Scans performed at limited intervals
    if (steamScanCnt >= 55) {
      await handleScan(scanSteamQuery, "Steam Market");
      steamScanCnt = 0;
    }
    if (csTradeScanCnt >= 100) {
      //All these are all at once, only done every 100 cycles
      await handleScan(scanCSTrade, "CS Trade");
      await handleScan(scanLootFarm, "Loot Farm");
      await handleScan(scanTradeIt, "Trade It");
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

async function handleScan(
  scan: () => Promise<DealNotification[]>,
  name: string,
) {
  const result = await fromThrowableAsync(scan, `Scan failed for ${name}`);
  if (result.isErr()) {
    handleError(result.error, name);
  }
}

