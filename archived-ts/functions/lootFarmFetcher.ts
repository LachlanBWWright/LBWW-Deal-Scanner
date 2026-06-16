import { LootFarmResponseSchema } from "./apiValidators.js";
import { fetchAndValidateJson } from "./fetchJson.js";
import { Agent } from "undici";

const lootFarmDispatcher = new Agent({
  connect: {
    family: 4,
    autoSelectFamily: false,
    timeout: 30000,
  },
});

export async function fetchLootFarmItems() {
  return fetchAndValidateJson(
    "https://loot.farm/botsInventory_730.json",
    LootFarmResponseSchema,
    { dispatcher: lootFarmDispatcher },
    "Failed to fetch and validate loot.farm items",
  );
}
