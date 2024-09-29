-- CreateTable
CREATE TABLE "Globals" (
    "id" TEXT NOT NULL,
    "BOT_CLIENT_ID" TEXT NOT NULL,
    "CASH_CONVERTERS" BOOLEAN NOT NULL,
    "CASH_CONVERTERS_CHANNEL_ID" TEXT NOT NULL,
    "CASH_CONVERTERS_ROLE_ID" TEXT NOT NULL,
    "COMMAND_PERMISSION_ROLE_ID" TEXT NOT NULL,
    "CS_CHANNEL_ID" TEXT NOT NULL,
    "CS_ITEMS" BOOLEAN NOT NULL,
    "CS_MARKET_CHANNEL_ID" TEXT NOT NULL,
    "CS_MARKET_ROLE_ID" TEXT NOT NULL,
    "CS_ROLE_ID" TEXT NOT NULL,
    "DISCORD_GUILD_ID" TEXT NOT NULL,
    "DISCORD_TOKEN" TEXT NOT NULL,
    "EBAY" BOOLEAN NOT NULL,
    "EBAY_CHANNEL_ID" TEXT NOT NULL,
    "EBAY_ROLE_ID" TEXT NOT NULL,
    "ERROR_CHANNEL_ID" TEXT NOT NULL,
    "GUMTREE" BOOLEAN NOT NULL,
    "GUMTREE_CHANNEL_ID" TEXT NOT NULL,
    "GUMTREE_ROLE_ID" TEXT NOT NULL,
    "SALVOS" BOOLEAN NOT NULL,
    "SALVOS_CHANNEL_ID" TEXT NOT NULL,
    "SALVOS_ROLE_ID" TEXT NOT NULL,
    "STEAM_QUERY" BOOLEAN NOT NULL,
    "STEAM_QUERY_CHANNEL_ID" TEXT NOT NULL,
    "STEAM_QUERY_ROLE_ID" TEXT NOT NULL,
    "CS_TRADE_DOLLAR_RATIO" REAL NOT NULL DEFAULT 1,
    "LOOT_FARM_DOLLAR_RATIO" REAL NOT NULL DEFAULT 1,
    "CS_DEALS_DOLLAR_RATIO" REAL NOT NULL DEFAULT 1,
    "TRADEIT_GG_DOLLAR_RATIO" REAL NOT NULL DEFAULT 1
);

-- CreateTable
CREATE TABLE "TtlItem" (
    "itemId" TEXT NOT NULL,
    "scanner" INTEGER NOT NULL,
    "lastUpdated" DATETIME NOT NULL,

    PRIMARY KEY ("scanner", "itemId")
);

-- CreateTable
CREATE TABLE "CashConverters" (
    "url" TEXT NOT NULL,
    "requiredPhrases" TEXT NOT NULL DEFAULT '',
    "excludePhrases" TEXT NOT NULL DEFAULT ''
);

-- CreateTable
CREATE TABLE "CsMarket" (
    "url" TEXT NOT NULL,
    "displayUrl" TEXT NOT NULL,
    "maxPrice" REAL NOT NULL,
    "maxFloat" REAL NOT NULL,
    "lastPrice" REAL NOT NULL
);

-- CreateTable
CREATE TABLE "CsTradeBot" (
    "name" TEXT NOT NULL,
    "maxPrice" REAL NOT NULL,
    "minFloat" REAL NOT NULL,
    "maxFloat" REAL NOT NULL
);

-- CreateTable
CREATE TABLE "Ebay" (
    "url" TEXT NOT NULL,
    "maxPrice" REAL NOT NULL
);

-- CreateTable
CREATE TABLE "Gumtree" (
    "url" TEXT NOT NULL,
    "maxPrice" REAL NOT NULL
);

-- CreateTable
CREATE TABLE "Salvos" (
    "name" TEXT NOT NULL,
    "minPrice" REAL NOT NULL,
    "maxPrice" REAL NOT NULL
);

-- CreateTable
CREATE TABLE "SteamMarket" (
    "name" TEXT NOT NULL,
    "displayUrl" TEXT NOT NULL,
    "maxPrice" REAL NOT NULL,
    "lastPrice" REAL NOT NULL
);

-- CreateIndex
CREATE UNIQUE INDEX "Globals_id_key" ON "Globals"("id");

-- CreateIndex
CREATE UNIQUE INDEX "TtlItem_itemId_key" ON "TtlItem"("itemId");

-- CreateIndex
CREATE UNIQUE INDEX "TtlItem_scanner_itemId_key" ON "TtlItem"("scanner", "itemId");

-- CreateIndex
CREATE UNIQUE INDEX "CashConverters_url_key" ON "CashConverters"("url");

-- CreateIndex
CREATE UNIQUE INDEX "CsMarket_url_key" ON "CsMarket"("url");

-- CreateIndex
CREATE UNIQUE INDEX "CsTradeBot_name_key" ON "CsTradeBot"("name");

-- CreateIndex
CREATE UNIQUE INDEX "Ebay_url_key" ON "Ebay"("url");

-- CreateIndex
CREATE UNIQUE INDEX "Gumtree_url_key" ON "Gumtree"("url");

-- CreateIndex
CREATE UNIQUE INDEX "Salvos_name_key" ON "Salvos"("name");

-- CreateIndex
CREATE UNIQUE INDEX "SteamMarket_name_key" ON "SteamMarket"("name");
