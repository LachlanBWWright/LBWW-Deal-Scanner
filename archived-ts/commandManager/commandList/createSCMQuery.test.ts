import { describe, expect, it } from "vitest";
import command, {
  deletescmquery,
  editscmquery,
  viewscmqueries,
} from "./createSCMQuery.js";

describe("commandList/createSCMQuery.ts", () => {
  it("defines createscmquery command schema", () => {
    const json = command.toJSON();
    expect(json.name).toBe("createscmquery");
    expect(json.options?.map((o: { name: string }) => o.name)).toEqual([
      "query",
      "maxprice",
      "dmonly",
    ]);
  });

  it("defines edit/delete/view command schemas", () => {
    expect(editscmquery.toJSON().name).toBe("editscmquery");
    expect(deletescmquery.toJSON().name).toBe("deletescmquery");
    expect(viewscmqueries.toJSON().name).toBe("viewscmqueries");
  });
});
