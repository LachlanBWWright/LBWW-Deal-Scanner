import { describe, expect, it, test } from "vitest";
import { getCsQueryString, getQueryResults } from "./steamMarket.js";

describe("steamMarket.ts", () => {
  test("fetches live steam render API results", async () => {
    const res = await getQueryResults(
      "https://steamcommunity.com/market/search/render/?query=ak-47&start=0&count=10&country=AU&language=english&currency=1&norender=1",
    );

    expect(Array.isArray(res)).toBe(true);
    if (res.length > 0) {
      expect(res[0]?.name).toBeTypeOf("string");
      expect(res[0]?.sell_price).toBeTypeOf("number");
      expect(res[0]?.asset_description?.appid).toBe(730);
    }
  }, 60000);

  it("builds a render API URL from a valid steam search url", async () => {
    const expectedResponseUrl =
      "https://steamcommunity.com/market/search/render/?query=ak-47&start=0&count=10&country=AU&language=english&currency=1";

    const fakeResponse = {
      url: () => expectedResponseUrl,
    };

    const fakePage = {
      goto: () => Promise.resolve(null),
      waitForResponse: async (
        predicate: (response: { url: () => string }) => boolean,
      ) => {
        if (!predicate(fakeResponse)) {
          throw new Error("Steam response predicate did not match");
        }
        return fakeResponse;
      },
      waitForNetworkIdle: () => Promise.resolve(),
    };

    const url = await getCsQueryString(
      fakePage,
      "https://steamcommunity.com/market/search?appid=730&q=ak-47#p1_price_asc",
    );

    expect(url).toBe(`${expectedResponseUrl}&norender=1`);
  });

  it("returns empty string for invalid steam search urls", async () => {
    const fakePage = {
      goto: () => Promise.resolve(null),
      waitForResponse: () => Promise.reject(new Error("should not be called")),
      waitForNetworkIdle: () => Promise.resolve(),
    };

    const url = await getCsQueryString(fakePage, "https://example.com");

    expect(url).toBe("");
  });
});
