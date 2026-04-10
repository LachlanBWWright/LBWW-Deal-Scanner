import puppeteer, { type Page } from "puppeteer";
import { memoryUsage } from "node:process";

import handleError from "./functions/handleError.js";
import { beginScanRun, beginScanLoop, endScanLoop, finishScanRun } from "./controlState.js";
import { scanCSTrade } from "./scanners/siteScanners/csTrade.js";
import { scanTradeIt } from "./scanners/siteScanners/tradeIt.js";
import { scanLootFarm } from "./scanners/siteScanners/lootFarm.js";
import { scanSteamQuery } from "./scanners/siteScanners/steamMarket.js";
import { scanCashConverters } from "./scanners/siteScanners/cashConverters.js";
import { scanSalvos } from "./scanners/siteScanners/salvos.js";
import { scanEbay } from "./scanners/siteScanners/ebay.js";
import { scanGumtree } from "./scanners/siteScanners/gumtree.js";

let loopPromise: Promise<void> | null = null;

export async function runScanOnce() {
  const record = beginScanRun("manual");
  const browser = await puppeteer.launch({
    args: ["--no-sandbox"],
    headless: true,
  });

  try {
    const page = await browser.newPage();
    await runScanPass(page, record.mode);
    await page.close();
    finishScanRun(record);
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown scan error";
    finishScanRun(record, message);
    throw error;
  } finally {
    await browser.close();
  }

  return record;
}

export function startBackgroundScanLoop() {
  if (loopPromise) return loopPromise;

  beginScanLoop();
  loopPromise = runLoop().finally(() => {
    endScanLoop();
    loopPromise = null;
  });

  return loopPromise;
}

async function runLoop() {
  const browser = await puppeteer.launch({
    args: ["--no-sandbox"],
    headless: true,
  });
  let page = await browser.newPage();
  let steamScanCnt = 0;
  let csTradeScanCnt = 0;

  try {
    while (true) {
      console.log(memoryUsage());
      const throttler = new Promise((resolve) => setTimeout(resolve, 2000));
      console.time("Cycle Time (2000ms minimum)");

      const record = beginScanRun("loop");
      try {
        await runScanPass(page, "loop", steamScanCnt, csTradeScanCnt);
        finishScanRun(record);
      } catch (error) {
        const message = error instanceof Error ? error.message : "Unknown scan error";
        finishScanRun(record, message);
      }

      steamScanCnt++;
      csTradeScanCnt++;

      await page.close();
      page = await browser.newPage();

      await throttler;
      console.timeEnd("Cycle Time (2000ms minimum)");
    }
  } finally {
    await page.close().catch(() => undefined);
    await browser.close().catch(() => undefined);
  }
}

async function runScanPass(
  page: Page,
  mode: "loop" | "manual",
  steamScanCnt = 0,
  csTradeScanCnt = 0,
) {
  if (mode === "loop") {
    await handleScan(() => scanCashConverters(page), "Cash Converters");
    await handleScan(() => scanEbay(page), "Ebay");
    await handleScan(() => scanGumtree(page), "Gumtree");
    await handleScan(() => scanSalvos(page), "Salvos");

    if (steamScanCnt >= 55) {
      await scanSteamQuery();
    }
    if (csTradeScanCnt >= 100) {
      await handleScan(scanCSTrade, "CS Trade");
      await handleScan(scanLootFarm, "Loot Farm");
      await handleScan(scanTradeIt, "Trade It");
    }
    return;
  }

  await handleScan(() => scanCashConverters(page), "Cash Converters");
  await handleScan(() => scanEbay(page), "Ebay");
  await handleScan(() => scanGumtree(page), "Gumtree");
  await handleScan(() => scanSalvos(page), "Salvos");
  await scanSteamQuery();
  await handleScan(scanCSTrade, "CS Trade");
  await handleScan(scanLootFarm, "Loot Farm");
  await handleScan(scanTradeIt, "Trade It");
}

async function handleScan(scan: () => Promise<void>, name: string) {
  try {
    await scan();
  } catch (error) {
    handleError(error, name);
  }
}
