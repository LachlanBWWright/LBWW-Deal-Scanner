import { describe, expect, it } from "vitest";
import command, {
  deletecsmarket,
  editcsmarket,
  viewcsmarketqueries,
} from "./createCSMarket.js";

describe("commandList/createCSMarket.ts", () => {
  it("defines createcsmarket command schema", () => {
    const json = command.toJSON();
    expect(json.name).toBe("createcsmarket");
    expect(json.options?.map((o: { name: string }) => o.name)).toEqual([
      "query",
      "maxfloat",
      "maxprice",
      "dmonly",
    ]);
  });

  it("defines edit/delete/view command schemas", () => {
    expect(editcsmarket.toJSON().name).toBe("editcsmarket");
    expect(deletecsmarket.toJSON().name).toBe("deletecsmarket");
    expect(viewcsmarketqueries.toJSON().name).toBe("viewcsmarketqueries");
  });
});
