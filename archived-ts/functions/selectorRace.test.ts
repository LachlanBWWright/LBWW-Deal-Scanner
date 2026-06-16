import { describe, expect, it } from "vitest";
import selectorRace from "./selectorRace.js";

describe("selectorRace.ts", () => {
  it("returns found element when item selector resolves first", async () => {
    const foundElement = { label: "found" };
    const page = {
      waitForSelector: async (selector: string) => {
        if (selector === ".item") {
          return foundElement;
        }
        return new Promise(() => undefined);
      },
    };

    const result = await selectorRace(page, ".item", ".empty", 100);

    expect(result).toEqual(foundElement);
  });

  it("returns null when no-item selector resolves first", async () => {
    const page = {
      waitForSelector: async (selector: string) => {
        if (selector === ".empty") {
          return null;
        }
        return new Promise(() => undefined);
      },
    };

    const result = await selectorRace(page, ".item", ".empty", 100);

    expect(result).toBeNull();
  });
});
