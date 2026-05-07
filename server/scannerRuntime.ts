import puppeteer, { type Page } from "puppeteer";
import { memoryUsage } from "node:process";

import handleError from "./functions/handleError.js";
import {
  beginScanRun,
  beginScanLoop,
  endScanLoop,
  finishScanRun,
  recordSearchResults,
} from "./controlState.js";
import { scanCSTrade } from "./scanners/siteScanners/csTrade.js";
import { scanTradeIt } from "./scanners/siteScanners/tradeIt.js";
import { scanLootFarm } from "./scanners/siteScanners/lootFarm.js";
import { scanSteamQuery } from "./scanners/siteScanners/steamMarket.js";
import { scanCashConverters } from "./scanners/siteScanners/cashConverters.js";
import { scanSalvos } from "./scanners/siteScanners/salvos.js";
import { scanEbay } from "./scanners/siteScanners/ebay.js";
import { scanGumtree } from "./scanners/siteScanners/gumtree.js";
import { fromThrowableAsync } from "./functions/neverthrowUtils.js";
import type { DealNotification } from "./deals/types.js";
import type { NotificationService } from "./notifications/types.js";
import { setSharedNotificationService } from "./notificationServiceRef.js";

let loopPromise: Promise<void> | null = null;
let notificationService: NotificationService | null = null;

export function setNotificationService(service: NotificationService) {
  notificationService = service;
  setSharedNotificationService(service);
}

export async function runScanOnce() {
  const record = beginScanRun("manual");
  const browserResult = await fromThrowableAsync(
    () =>
      puppeteer.launch({
        args: ["--no-sandbox"],
        headless: true,
      }),
    "Failed to launch scan browser",
  );
  if (browserResult.isErr()) {
    finishScanRun(record, browserResult.error.message);
    throw browserResult.error;
  }
  const browser = browserResult.value;

  const pageResult = await fromThrowableAsync(
    () => browser.newPage(),
    "Failed to create scan page",
  );
  if (pageResult.isErr()) {
    await fromThrowableAsync(
      () => browser.close(),
      "Failed to close scan browser",
    );
    finishScanRun(record, pageResult.error.message);
    throw pageResult.error;
  }
  const page = pageResult.value;

  const scanResult = await fromThrowableAsync(
    () => runScanPass(page, record.mode),
    "Scan pass failed",
  );
  await fromThrowableAsync(() => page.close(), "Failed to close scan page");
  await fromThrowableAsync(
    () => browser.close(),
    "Failed to close scan browser",
  );

  if (scanResult.isErr()) {
    finishScanRun(record, scanResult.error.message);
    throw scanResult.error;
  }

  finishScanRun(record);

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
  const browserResult = await fromThrowableAsync(
    () =>
      puppeteer.launch({
        args: ["--no-sandbox"],
        headless: true,
      }),
    "Failed to launch scan loop browser",
  );
  if (browserResult.isErr()) {
    handleError(
      browserResult.error,
      "Scanner Runtime",
      notificationService ?? undefined,
    );
    return;
  }
  const browser = browserResult.value;

  const firstPageResult = await fromThrowableAsync(
    () => browser.newPage(),
    "Failed to create scan loop page",
  );
  if (firstPageResult.isErr()) {
    handleError(
      firstPageResult.error,
      "Scanner Runtime",
      notificationService ?? undefined,
    );
    await fromThrowableAsync(
      () => browser.close(),
      "Failed to close scan loop browser",
    );
    return;
  }
  let page = firstPageResult.value;
  let steamScanCnt = 0;
  let csTradeScanCnt = 0;

  while (true) {
    console.log(memoryUsage());
    const throttler = new Promise((resolve) => setTimeout(resolve, 2000));
    console.time("Cycle Time (2000ms minimum)");

    const record = beginScanRun("loop");
    const scanResult = await fromThrowableAsync(
      () => runScanPass(page, "loop", steamScanCnt, csTradeScanCnt),
      "Scan pass failed",
    );
    if (scanResult.isErr()) {
      finishScanRun(record, scanResult.error.message);
    } else {
      finishScanRun(record);
    }

    steamScanCnt++;
    csTradeScanCnt++;

    await fromThrowableAsync(
      () => page.close(),
      "Failed to close scan loop page",
    );
    const nextPageResult = await fromThrowableAsync(
      () => browser.newPage(),
      "Failed to create scan loop page",
    );
    if (nextPageResult.isErr()) {
      handleError(
        nextPageResult.error,
        "Scanner Runtime",
        notificationService ?? undefined,
      );
      await fromThrowableAsync(
        () => browser.close(),
        "Failed to close scan loop browser",
      );
      return;
    }
    page = nextPageResult.value;

    await throttler;
    console.timeEnd("Cycle Time (2000ms minimum)");
  }
}

async function publishNotifications(notifications: DealNotification[]) {
  recordSearchResults(
    notifications.map((notification) => ({
      source: notification.source,
      title: notification.title,
      url: notification.url,
      price: notification.price ?? null,
      imageUrl: notification.imageUrl ?? null,
      queryType: notification.query?.type ?? null,
      queryId: notification.query?.id ?? null,
      foundAt: new Date().toISOString(),
    })),
  );

  if (!notificationService) return;
  for (const notification of notifications) {
    await notificationService.publish(notification);
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
      await handleScan(scanSteamQuery, "Steam Market");
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
  await handleScan(scanSteamQuery, "Steam Market");
  await handleScan(scanCSTrade, "CS Trade");
  await handleScan(scanLootFarm, "Loot Farm");
  await handleScan(scanTradeIt, "Trade It");
}

async function handleScan(
  scan: () => Promise<DealNotification[]>,
  name: string,
) {
  const result = await fromThrowableAsync(scan, `Scan failed for ${name}`);
  if (result.isErr()) {
    handleError(result.error, name, notificationService ?? undefined);
  } else {
    await publishNotifications(result.value);
  }
}
