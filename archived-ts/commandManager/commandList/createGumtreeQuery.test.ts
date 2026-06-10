import { describe, expect, it } from "vitest";
import command, {
  deletegumtreequery,
  editgumtreequery,
  viewgumtreequeries,
} from "./createGumtreeQuery.js";

describe("commandList/createGumtreeQuery.ts", () => {
  it("defines creategumtreequery command schema", () => {
    const json = command.toJSON();
    expect(json.name).toBe("creategumtreequery");
    expect(json.options?.map((o: { name: string }) => o.name)).toEqual([
      "query",
      "maxprice",
      "dmonly",
    ]);
  });

  it("defines edit/delete/view command schemas", () => {
    expect(editgumtreequery.toJSON().name).toBe("editgumtreequery");
    expect(deletegumtreequery.toJSON().name).toBe("deletegumtreequery");
    expect(viewgumtreequeries.toJSON().name).toBe("viewgumtreequeries");
  });
});
