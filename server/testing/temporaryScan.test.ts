import { okAsync } from "neverthrow";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { runTemporaryScan } from "./temporaryScan.js";

const { fetchCsTradeItems } = vi.hoisted(() => ({
  fetchCsTradeItems: vi.fn(),
}));

vi.mock("../functions/csTraceFetcher.js", () => ({
  fetchCsTradeItems,
}));

describe("temporaryScan.ts", () => {
  beforeEach(() => {
    fetchCsTradeItems.mockReset();
  });

  it("runs csTradeBot scans without persisted queries", async () => {
    fetchCsTradeItems.mockReturnValue(
      okAsync([
        {
          id: "1",
          app_id: 730,
          market_hash_name: "AK-47 | Redline (Field-Tested)",
          price: 42,
          wear: 0.22,
        },
        {
          id: "2",
          app_id: 730,
          market_hash_name: "AK-47 | Redline (Field-Tested)",
          price: 80,
          wear: 0.22,
        },
      ]),
    );

    const result = await runTemporaryScan("csTradeBot", {
      name: "AK-47 | Redline (Field-Tested)",
      minFloat: 0.15,
      maxFloat: 0.3,
      maxPrice: 50,
    });

    expect(result.errors).toEqual([]);
    expect(result.notifications).toHaveLength(1);
    expect(result.notifications[0]?.query).toEqual({
      type: "csTradeBot",
      id: "AK-47 | Redline (Field-Tested)",
    });
  });

  it("validates required csTradeBot numeric fields before scanning", async () => {
    const result = await runTemporaryScan("csTradeBot", {
      name: "AK-47 | Redline (Field-Tested)",
      minFloat: "",
      maxFloat: 0.3,
      maxPrice: 50,
    });

    expect(fetchCsTradeItems).not.toHaveBeenCalled();
    expect(result.notifications).toEqual([]);
    expect(result.errors).toEqual(["minFloat is required"]);
  });
});
