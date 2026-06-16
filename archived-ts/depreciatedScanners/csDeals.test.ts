import { describe, expect, it } from "vitest";
import { getCSDealsItems } from "./csDeals.js";

describe("depreciatedScanners/csDeals.ts", () => {
  it("extracts response payload from botsinventory endpoint", async () => {
    const expected = [{ c: "AK-47", d1: 0.12, i: 4.2 }];
    const response = {
      url: () => "https://api.cs.deals/botsinventory?appid=0",
      json: async () => ({ response: expected }),
    };

    const page = {
      setDefaultNavigationTimeout: () => undefined,
      goto: async () => undefined,
      waitForResponse: async (
        predicate: (candidate: { url(): string }) => boolean,
      ) => {
        if (predicate(response)) {
          return response;
        }
        return null;
      },
    };

    const result = await getCSDealsItems(page);
    expect(result).toEqual(expected);
  });
});
