import { describe, it, expect } from "vitest";
import { formatPrice } from "../functions/formatPrice.js";

describe("formatPrice", () => {
  it("should format positive numbers as AUD currency", () => {
    expect(formatPrice(10.50)).toBe("$10.50");
    expect(formatPrice(1000)).toBe("$1,000.00");
    expect(formatPrice(1234.56)).toBe("$1,234.56");
  });

  it("should format zero as currency", () => {
    expect(formatPrice(0)).toBe("$0.00");
  });

  it("should format negative numbers as currency", () => {
    expect(formatPrice(-10.50)).toBe("-$10.50");
    expect(formatPrice(-1000)).toBe("-$1,000.00");
  });

  it("should handle decimal precision correctly", () => {
    expect(formatPrice(9.99)).toBe("$9.99");
    expect(formatPrice(9.999)).toBe("$10.00");
    expect(formatPrice(9.994)).toBe("$9.99");
  });

  it("should handle large numbers", () => {
    expect(formatPrice(1000000)).toBe("$1,000,000.00");
    expect(formatPrice(1234567.89)).toBe("$1,234,567.89");
  });
});