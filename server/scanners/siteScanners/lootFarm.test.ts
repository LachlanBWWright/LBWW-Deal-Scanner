import { describe, expect, it, test } from "vitest";
import { fetchLootFarmItems } from "../../functions/lootFarmFetcher.js";

test("loot farm scanner", async () => {
  const result = await fetchLootFarmItems();

  expect(result.isOk()).toBe(true);
  if (result.isErr()) throw result.error;

  expect(Object.keys(result.value.result).length).toBeGreaterThan(0);
}, 60000);

describe("lootFarm.ts", () => {
  it("exports a live fetcher", () => {
    expect(fetchLootFarmItems).toBeTypeOf("function");
  });
});

describe("lootFarm.ts", () => {
  it("returns skins with string names from the validated payload", async () => {
    const result = await fetchLootFarmItems();
    expect(result.isOk()).toBe(true);
    if (result.isErr()) throw result.error;

    const firstSkin = Object.values(result.value.result)[0];
    expect(firstSkin?.n).toBeTypeOf("string");
  });
});
