import puppeteer from "puppeteer";
import { expect, test } from "vitest";
import { getCsQueryString, getQueryResults } from "./steamMarket";

test("steam query scanner", async () => {
  const browser = await puppeteer.launch({
    headless: "shell",
    args: ["--no-sandbox"],
  });
  const page = await browser.newPage();

  const url = await getCsQueryString(
    page,
    "https://steamcommunity.com/market/search?appid=730&q=test#p1_price_asc",
  );
  console.log(url);

  await page.close();
  await browser.close();

  const res = await getQueryResults(url);

  console.log(res);
});

import { describe, it } from "vitest";
describe("steamMarket.ts", () => {
  it("returns empty string for invalid steam search urls", async () => {
    const browser = await puppeteer.launch({
      headless: "shell",
      args: ["--no-sandbox"],
    });
    const page = await browser.newPage();
    const url = await getCsQueryString(page, "https://example.com");
    await page.close();
    await browser.close();

    expect(url).toBe("");
  });
});
