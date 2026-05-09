import { describe, expect, it } from "vitest";
import command, {
  deletecashquery,
  editcashquery,
  viewcashqueries,
} from "./createCashQuery.js";

describe("commandList/createCashQuery.ts", () => {
  it("defines createcashquery command schema", () => {
    const json = command.toJSON();
    expect(json.name).toBe("createcashquery");
    expect(json.options?.map((o: { name: string }) => o.name)).toEqual([
      "query",
      "dmonly",
    ]);
  });

  it("defines edit/delete/view command schemas", () => {
    expect(editcashquery.toJSON().name).toBe("editcashquery");
    expect(deletecashquery.toJSON().name).toBe("deletecashquery");
    expect(viewcashqueries.toJSON().name).toBe("viewcashqueries");
  });
});
