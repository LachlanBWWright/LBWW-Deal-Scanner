import { describe, expect, it } from "vitest";
import { commandHandler, commandList } from "./index.js";

describe("commandManager/index.ts", () => {
  it("registers expected slash commands", () => {
    const names = commandList.map((command) => command.toJSON().name);
    expect(names).toContain("createcashquery");
    expect(names).toContain("createebayquery");
    expect(names).toContain("createscmquery");
    expect(commandList.length).toBeGreaterThan(20);
  });

  it("exposes a callable command handler", () => {
    expect(typeof commandHandler).toBe("function");
  });
});
