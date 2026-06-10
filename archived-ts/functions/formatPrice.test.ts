import { describe, expect, it } from "vitest";
import { formatPrice } from "./formatPrice.js";

describe("formatPrice.ts", () => {
  it("formats AUD currency with cents precision", () => {
    expect(formatPrice(10)).toBe("$10.00");
    expect(formatPrice(10.5)).toBe("$10.50");
  });
});
