import { describe, it } from "vitest";
describe("salvos.ts", () => {
  it("should ...", () => {
    // TODO: implement test
  });
});
import { chromium } from "playwright";
import { test } from "vitest";
import { getSalvosValues } from "./salvos";

test("salvos scanner", async () => {
  const browser = await chromium.launch({
    headless: true,
    args: ["--no-sandbox"],
  });
  let page = await browser.newPage();
  const results = await getSalvosValues(page, {
    name: "shirt",
    minPrice: 0,
    maxPrice: 50,
  });
  console.log(results);
});
