-- CreateTable
CREATE TABLE "ActionRegistry" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "type" TEXT NOT NULL,
    "queryType" TEXT,
    "queryId" TEXT,
    "userId" TEXT,
    "timestamp" BIGINT NOT NULL,
    "relatedKey" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
