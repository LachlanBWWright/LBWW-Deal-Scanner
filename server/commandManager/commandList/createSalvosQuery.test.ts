import { describe, expect, it } from "vitest";
import command, {
  deletesalvosquery,
  editsalvosquery,
  viewsalvosqueries,
} from "./createSalvosQuery.js";

describe("commandList/createSalvosQuery.ts", () => {
  it("defines createsalvosquery command schema", () => {
    const json = command.toJSON();
    expect(json.name).toBe("createsalvosquery");
    expect(json.options?.map((o: { name: string }) => o.name)).toEqual([
      "query",
      "minprice",
      "maxprice",
      "dmonly",
    ]);
  });

  it("defines edit/delete/view command schemas", () => {
    expect(editsalvosquery.toJSON().name).toBe("editsalvosquery");
    expect(deletesalvosquery.toJSON().name).toBe("deletesalvosquery");
    expect(viewsalvosqueries.toJSON().name).toBe("viewsalvosqueries");
  });
});
