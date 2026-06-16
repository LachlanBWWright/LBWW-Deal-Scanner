const displayLabels: Record<string, string> = {
  cashConverters: "Cash Converters",
  csMarket: "CS Market",
  csTrade: "CS Trade",
  csTradeBot: "CS Trade Bot",
  ebay: "eBay",
  gumtree: "Gumtree",
  lootFarm: "Loot Farm",
  salvos: "Salvos",
  steamMarket: "Steam Market",
  tradeIt: "Trade It",
  TestScanner: "Test Scanner",
};

export function formatDisplayLabel(value: string): string {
  const mapped = displayLabels[value];
  if (mapped) {
    return mapped;
  }

  return value
    .replace(/([a-z0-9])([A-Z])/g, "$1 $2")
    .replace(/[_-]+/g, " ")
    .replace(/\b\w/g, (character) => character.toUpperCase());
}
