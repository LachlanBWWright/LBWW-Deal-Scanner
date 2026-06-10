import { describe, expect, it } from "vitest";
import scannerLoop from "./index.js";

describe("scanners/index.ts", () => {
  it("exports an async scanner runtime loop function", () => {
    expect(typeof scannerLoop).toBe("function");
    expect(scannerLoop.constructor.name).toBe("AsyncFunction");
  });
});
