import puppeteer, { type Page } from "puppeteer";
import { memoryUsage } from "node:process";

import handleError from "./functions/handleError.js";
import { db } from "./globals/PrismaClient.js";
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
import { resultAsync } from "./functions/neverthrowUtils.js";
import type { DealNotification } from "./deals/types.js";
import type { NotificationService } from "./notifications/types.js";
import { setSharedNotificationService } from "./notificationServiceRef.js";

let loopPromise: Promise<void> | null = null;
let notificationService: NotificationService | null = null;
const RUNTIME_STATE_ID = "scanner-runtime";

export function setNotificationService(service: NotificationService) {
  notificationService = service;
  setSharedNotificationService(service);
}

export async function runScanOnce() {
  const record = beginScanRun("manual");
  const browserResult = await resultAsync(
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

  const pageResult = await resultAsync(
    () => browser.newPage(),
    "Failed to create scan page",
  );
  if (pageResult.isErr()) {
    await resultAsync(
      () => browser.close(),
      "Failed to close scan browser",
    );
    finishScanRun(record, pageResult.error.message);
    throw pageResult.error;
  }
  const page = pageResult.value;

  const scanResult = await resultAsync(
    () => runScanPass(page, record.mode),
    "Scan pass failed",
  );
  await resultAsync(() => page.close(), "Failed to close scan page");
  await resultAsync(
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
  loopPromise = runLoop(null).finally(() => {
    endScanLoop();
    loopPromise = null;
  });

  return loopPromise;
}

export async function runScheduledScanLoop(durationMs: number) {
  beginScanLoop();
  await markScheduledRunStarted();
  const startedAt = Date.now();
  await runLoop(durationMs);
  const duration = Date.now() - startedAt;
  await markScheduledRunStopped(duration);
  endScanLoop();
}

async function runLoop(maxDurationMs: number | null) {
  const browserResult = await resultAsync(
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

  const firstPageResult = await resultAsync(
    () => browser.newPage(),
    "Failed to create scan loop page",
  );
  if (firstPageResult.isErr()) {
    handleError(
      firstPageResult.error,
      "Scanner Runtime",
      notificationService ?? undefined,
    );
    await resultAsync(
      () => browser.close(),
      "Failed to close scan loop browser",
    );
    return;
  }
  let page: Page | null = firstPageResult.value;
  const persistedState = await readScannerRuntimeState();
  let steamScanCnt = persistedState.steamScanCount;
  let csTradeScanCnt = persistedState.csTradeScanCount;
  const loopStartedAt = Date.now();

  while (maxDurationMs === null || Date.now() - loopStartedAt < maxDurationMs) {
    console.log(memoryUsage());
    const throttler = new Promise((resolve) => setTimeout(resolve, 2000));
    console.time("Cycle Time (2000ms minimum)");

    const record = beginScanRun("loop");
    const currentPage = page;
    if (currentPage === null) {
      break;
    }
    const scannerState = {
      steamScanCnt,
      csTradeScanCnt,
    };
    const scanResult = await resultAsync(
      () =>
        runScanPass(
          currentPage,
          "loop",
          scannerState.steamScanCnt,
          scannerState.csTradeScanCnt,
        ),
      "Scan pass failed",
    );
    if (scanResult.isErr()) {
      finishScanRun(record, scanResult.error.message);
    } else {
      finishScanRun(record);
    }

    steamScanCnt = scannerState.steamScanCnt >= 55 ? 0 : scannerState.steamScanCnt + 1;
    csTradeScanCnt =
      scannerState.csTradeScanCnt >= 100 ? 0 : scannerState.csTradeScanCnt + 1;
    await writeScannerCounters(steamScanCnt, csTradeScanCnt);

    await resultAsync(
      () => currentPage.close(),
      "Failed to close scan loop page",
    );
    page = null;

    if (maxDurationMs !== null && Date.now() - loopStartedAt >= maxDurationMs) {
      console.timeEnd("Cycle Time (2000ms minimum)");
      break;
    }

    const nextPageResult = await resultAsync(
      () => browser.newPage(),
      "Failed to create scan loop page",
    );
    if (nextPageResult.isErr()) {
      handleError(
        nextPageResult.error,
        "Scanner Runtime",
        notificationService ?? undefined,
      );
      await resultAsync(
        () => browser.close(),
        "Failed to close scan loop browser",
      );
      return;
    }
    page = nextPageResult.value;

    await throttler;
    console.timeEnd("Cycle Time (2000ms minimum)");
  }

  if (page !== null) {
    await resultAsync(() => page.close(), "Failed to close scan loop page");
  }
  await resultAsync(
    () => browser.close(),
    "Failed to close scan loop browser",
  );
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
  const result = await resultAsync(scan, `Scan failed for ${name}`);
  if (result.isErr()) {
    handleError(result.error, name, notificationService ?? undefined);
  } else {
    await publishNotifications(result.value);
  }
}

async function readScannerRuntimeState() {
  const stateResult = await resultAsync(
    () =>
      db.scannerRuntimeState.upsert({
        where: { id: RUNTIME_STATE_ID },
        create: { id: RUNTIME_STATE_ID },
        update: {},
      }),
    "Failed to read scanner runtime state",
  );

  if (stateResult.isErr()) {
    handleError(
      stateResult.error,
      "Scanner Runtime",
      notificationService ?? undefined,
    );
    return {
      steamScanCount: 0,
      csTradeScanCount: 0,
    };
  }

  return stateResult.value;
}

async function writeScannerCounters(
  steamScanCount: number,
  csTradeScanCount: number,
) {
  const stateResult = await resultAsync(
    () =>
      db.scannerRuntimeState.upsert({
        where: { id: RUNTIME_STATE_ID },
        create: {
          id: RUNTIME_STATE_ID,
          steamScanCount,
          csTradeScanCount,
        },
        update: {
          steamScanCount,
          csTradeScanCount,
        },
      }),
    "Failed to persist scanner runtime counters",
  );

  if (stateResult.isErr()) {
    handleError(
      stateResult.error,
      "Scanner Runtime",
      notificationService ?? undefined,
    );
  }
}

async function markScheduledRunStarted() {
  const stateResult = await resultAsync(
    () =>
      db.scannerRuntimeState.upsert({
        where: { id: RUNTIME_STATE_ID },
        create: {
          id: RUNTIME_STATE_ID,
          lastStartedAt: new Date(),
        },
        update: {
          lastStartedAt: new Date(),
        },
      }),
    "Failed to record scheduled scanner start",
  );

  if (stateResult.isErr()) {
    handleError(
      stateResult.error,
      "Scanner Runtime",
      notificationService ?? undefined,
    );
  }
}

async function markScheduledRunStopped(durationMs: number) {
  const stateResult = await resultAsync(
    () =>
      db.scannerRuntimeState.upsert({
        where: { id: RUNTIME_STATE_ID },
        create: {
          id: RUNTIME_STATE_ID,
          scheduledRuns: 1,
          totalScheduledRuntimeMs: durationMs,
          lastStoppedAt: new Date(),
        },
        update: {
          scheduledRuns: { increment: 1 },
          totalScheduledRuntimeMs: { increment: durationMs },
          lastStoppedAt: new Date(),
        },
      }),
    "Failed to record scheduled scanner stop",
  );

  if (stateResult.isErr()) {
    handleError(
      stateResult.error,
      "Scanner Runtime",
      notificationService ?? undefined,
    );
  }
}
