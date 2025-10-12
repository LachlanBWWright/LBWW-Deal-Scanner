-- CreateTable
CREATE TABLE "Query" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "dmOnly" BOOLEAN NOT NULL DEFAULT false,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "UserQuery" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "userId" TEXT NOT NULL,
    "queryId" TEXT NOT NULL,
    "queryType" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "UserQuery_queryId_fkey" FOREIGN KEY ("queryId") REFERENCES "Query" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- RedefineTables
PRAGMA defer_foreign_keys=ON;
PRAGMA foreign_keys=OFF;
CREATE TABLE "new_CashConverters" (
    "url" TEXT NOT NULL,
    "requiredPhrases" TEXT NOT NULL DEFAULT '',
    "excludePhrases" TEXT NOT NULL DEFAULT '',
    "queryId" TEXT,
    CONSTRAINT "CashConverters_queryId_fkey" FOREIGN KEY ("queryId") REFERENCES "Query" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);
INSERT INTO "new_CashConverters" ("excludePhrases", "requiredPhrases", "url") SELECT "excludePhrases", "requiredPhrases", "url" FROM "CashConverters";
DROP TABLE "CashConverters";
ALTER TABLE "new_CashConverters" RENAME TO "CashConverters";
CREATE UNIQUE INDEX "CashConverters_url_key" ON "CashConverters"("url");
CREATE UNIQUE INDEX "CashConverters_queryId_key" ON "CashConverters"("queryId");
CREATE TABLE "new_CsMarket" (
    "url" TEXT NOT NULL,
    "displayUrl" TEXT NOT NULL,
    "maxPrice" REAL NOT NULL,
    "maxFloat" REAL NOT NULL,
    "lastPrice" REAL NOT NULL,
    "queryId" TEXT,
    CONSTRAINT "CsMarket_queryId_fkey" FOREIGN KEY ("queryId") REFERENCES "Query" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);
INSERT INTO "new_CsMarket" ("displayUrl", "lastPrice", "maxFloat", "maxPrice", "url") SELECT "displayUrl", "lastPrice", "maxFloat", "maxPrice", "url" FROM "CsMarket";
DROP TABLE "CsMarket";
ALTER TABLE "new_CsMarket" RENAME TO "CsMarket";
CREATE UNIQUE INDEX "CsMarket_url_key" ON "CsMarket"("url");
CREATE UNIQUE INDEX "CsMarket_queryId_key" ON "CsMarket"("queryId");
CREATE TABLE "new_CsTradeBot" (
    "name" TEXT NOT NULL,
    "maxPrice" REAL NOT NULL,
    "minFloat" REAL NOT NULL,
    "maxFloat" REAL NOT NULL,
    "queryId" TEXT,
    CONSTRAINT "CsTradeBot_queryId_fkey" FOREIGN KEY ("queryId") REFERENCES "Query" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);
INSERT INTO "new_CsTradeBot" ("maxFloat", "maxPrice", "minFloat", "name") SELECT "maxFloat", "maxPrice", "minFloat", "name" FROM "CsTradeBot";
DROP TABLE "CsTradeBot";
ALTER TABLE "new_CsTradeBot" RENAME TO "CsTradeBot";
CREATE UNIQUE INDEX "CsTradeBot_name_key" ON "CsTradeBot"("name");
CREATE UNIQUE INDEX "CsTradeBot_queryId_key" ON "CsTradeBot"("queryId");
CREATE TABLE "new_Ebay" (
    "url" TEXT NOT NULL,
    "maxPrice" REAL NOT NULL,
    "queryId" TEXT,
    CONSTRAINT "Ebay_queryId_fkey" FOREIGN KEY ("queryId") REFERENCES "Query" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);
INSERT INTO "new_Ebay" ("maxPrice", "url") SELECT "maxPrice", "url" FROM "Ebay";
DROP TABLE "Ebay";
ALTER TABLE "new_Ebay" RENAME TO "Ebay";
CREATE UNIQUE INDEX "Ebay_url_key" ON "Ebay"("url");
CREATE UNIQUE INDEX "Ebay_queryId_key" ON "Ebay"("queryId");
CREATE TABLE "new_Gumtree" (
    "url" TEXT NOT NULL,
    "maxPrice" REAL NOT NULL,
    "queryId" TEXT,
    CONSTRAINT "Gumtree_queryId_fkey" FOREIGN KEY ("queryId") REFERENCES "Query" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);
INSERT INTO "new_Gumtree" ("maxPrice", "url") SELECT "maxPrice", "url" FROM "Gumtree";
DROP TABLE "Gumtree";
ALTER TABLE "new_Gumtree" RENAME TO "Gumtree";
CREATE UNIQUE INDEX "Gumtree_url_key" ON "Gumtree"("url");
CREATE UNIQUE INDEX "Gumtree_queryId_key" ON "Gumtree"("queryId");
CREATE TABLE "new_Salvos" (
    "name" TEXT NOT NULL,
    "minPrice" REAL NOT NULL,
    "maxPrice" REAL NOT NULL,
    "queryId" TEXT,
    CONSTRAINT "Salvos_queryId_fkey" FOREIGN KEY ("queryId") REFERENCES "Query" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);
INSERT INTO "new_Salvos" ("maxPrice", "minPrice", "name") SELECT "maxPrice", "minPrice", "name" FROM "Salvos";
DROP TABLE "Salvos";
ALTER TABLE "new_Salvos" RENAME TO "Salvos";
CREATE UNIQUE INDEX "Salvos_name_key" ON "Salvos"("name");
CREATE UNIQUE INDEX "Salvos_queryId_key" ON "Salvos"("queryId");
CREATE TABLE "new_SteamMarket" (
    "name" TEXT NOT NULL,
    "displayUrl" TEXT NOT NULL,
    "maxPrice" REAL NOT NULL,
    "lastPrice" REAL NOT NULL,
    "queryId" TEXT,
    CONSTRAINT "SteamMarket_queryId_fkey" FOREIGN KEY ("queryId") REFERENCES "Query" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);
INSERT INTO "new_SteamMarket" ("displayUrl", "lastPrice", "maxPrice", "name") SELECT "displayUrl", "lastPrice", "maxPrice", "name" FROM "SteamMarket";
DROP TABLE "SteamMarket";
ALTER TABLE "new_SteamMarket" RENAME TO "SteamMarket";
CREATE UNIQUE INDEX "SteamMarket_name_key" ON "SteamMarket"("name");
CREATE UNIQUE INDEX "SteamMarket_queryId_key" ON "SteamMarket"("queryId");
PRAGMA foreign_keys=ON;
PRAGMA defer_foreign_keys=OFF;

-- CreateIndex
CREATE UNIQUE INDEX "UserQuery_userId_queryId_key" ON "UserQuery"("userId", "queryId");
