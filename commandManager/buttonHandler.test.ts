import { describe, it, expect } from "vitest";
describe("buttonHandler.ts", () => {
  it("should export buttonInteractionHandler function", async () => {
    // Test that the module can be imported and exports the expected function
    const module = await import("./buttonHandler.js");
    expect(typeof module.buttonInteractionHandler).toBe("function");
  });
});
