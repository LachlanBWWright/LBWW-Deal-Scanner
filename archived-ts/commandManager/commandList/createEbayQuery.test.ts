import { describe, expect, it } from "vitest";
import command, {
  deleteebayquery,
  editedbayquery,
  viewebayqueries,
} from "./createEbayQuery.js";

describe("commandList/createEbayQuery.ts", () => {
  it("defines createebayquery command schema", () => {
    const json = command.toJSON();
    expect(json.name).toBe("createebayquery");
    expect(json.options?.map((o: { name: string }) => o.name)).toEqual([
      "query",
      "maxprice",
      "dmonly",
    ]);
  });

  it("defines edit/delete/view command schemas", () => {
    expect(editedbayquery.toJSON().name).toBe("editedbayquery");
    expect(deleteebayquery.toJSON().name).toBe("deleteebayquery");
    expect(viewebayqueries.toJSON().name).toBe("viewebayqueries");
  });
});
