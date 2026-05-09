import { describe, expect, it } from "vitest";
import globals from "./Globals.js";

describe("Globals.ts", () => {
  it("provides expected configuration keys used by scanners and routing", () => {
    expect(typeof globals).toBe("object");
    expect("EBAY" in globals).toBe(true);
    expect("GUMTREE" in globals).toBe(true);
    expect("SALVOS" in globals).toBe(true);
    expect("CS_ITEMS" in globals).toBe(true);
    expect("COMMAND_PERMISSION_ROLE_ID" in globals).toBe(true);
    expect("ERROR_CHANNEL_ID" in globals).toBe(true);
  });
});
