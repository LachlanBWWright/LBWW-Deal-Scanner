import { describe, expect, it, test } from "vitest";
import { fetchLootFarmItems } from "../../functions/lootFarmFetcher.js";

test("loot farm scanner", async () => {
  const result = await fetchLootFarmItems();

  if (result.isErr()) {
    expect(result.error.message).toBe("");
    return;
  }

  expect(Object.keys(result.value.result).length).toBeGreaterThan(0);
});

describe("lootFarm.ts", () => {
  it("exports a live fetcher", () => {
    expect(fetchLootFarmItems).toBeTypeOf("function");
  });
});

describe("lootFarm.ts", () => {
  it("should fetch and validate items", () => {
    // TODO: implement test
  });
});
