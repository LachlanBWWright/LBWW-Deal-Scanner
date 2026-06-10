import { describe, expect, it } from "vitest";
import command, {
  deletemultisearchquery,
  editmultisearchquery,
  viewmultisearchqueries,
} from "./createMultiSearchQuery.js";

describe("commandList/createMultiSearchQuery.ts", () => {
  it("defines createmultisearch command schema", () => {
    const json = command.toJSON();
    expect(json.name).toBe("createmultisearch");
    expect(json.options?.map((o: { name: string }) => o.name)).toEqual([
      "skinname",
      "minfloat",
      "maxfloat",
      "maxprice",
      "dmonly",
    ]);
  });

  it("defines edit/delete/view command schemas", () => {
    expect(editmultisearchquery.toJSON().name).toBe("editmultisearchquery");
    expect(deletemultisearchquery.toJSON().name).toBe("deletemultisearchquery");
    expect(viewmultisearchqueries.toJSON().name).toBe("viewmultisearchqueries");
  });
});
