-- CreateTable
CREATE TABLE "ScannerRuntimeState" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "steamScanCount" INTEGER NOT NULL DEFAULT 0,
    "csTradeScanCount" INTEGER NOT NULL DEFAULT 0,
    "scheduledRuns" INTEGER NOT NULL DEFAULT 0,
    "totalScheduledRuntimeMs" BIGINT NOT NULL DEFAULT 0,
    "lastStartedAt" DATETIME,
    "lastStoppedAt" DATETIME,
    "updatedAt" DATETIME NOT NULL
);
