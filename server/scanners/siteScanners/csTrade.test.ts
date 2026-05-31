import { describe, expect, test } from "vitest";
import { getCsTradeItems } from "./csTrade.js";

const runLiveApiTests = process.env.RUN_LIVE_API_TESTS === "true";

test.skipIf(!runLiveApiTests)(
  "cs Trade scanner fetches live item inventory",
  async () => {
    const result = await getCsTradeItems();

    expect(Array.isArray(result)).toBe(true);
    expect(result.length).toBeGreaterThan(0);
    expect(result[0]?.app_id).toBe(730);
    expect(result[0]?.market_hash_name).toBeTypeOf("string");
    expect(result[0]?.price).toBeTypeOf("number");
    expect(result[0]?.wear).toBeTypeOf("number");
  },
  60000,
);

describe("csTrade.ts", () => {
  test("exports live inventory function", () => {
    expect(getCsTradeItems).toBeTypeOf("function");
  });
});
