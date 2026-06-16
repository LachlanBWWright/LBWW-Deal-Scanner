import puppeteer from "puppeteer";
import { describe, expect, test } from "vitest";
import { getGumtreeValues } from "./gumtree.js";

const runLiveApiTests = process.env.RUN_LIVE_API_TESTS === "true";

test.skipIf(!runLiveApiTests)(
  "gumtree scanner performs live scrape without crashing",
  async () => {
  const browser = await puppeteer.launch({
    args: [
      "--no-sandbox",
      "--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
    ],
  });

  const page = await browser.newPage();
  const result = await getGumtreeValues(page, {
    url: "https://www.gumtree.com.au/s-playstation/ballarat-city/ps4/k0c18617l3000626r10",
  });

  // Gumtree markup changes frequently; null is accepted but undefined is not.
  expect(result === null || typeof result === "object").toBe(true);
  if (result) {
    expect(result.foundName).toBeTypeOf("string");
    expect(result.foundPrice).toBeTypeOf("number");
    expect(Number.isFinite(result.foundPrice)).toBe(true);
  }

  await browser.close();
  },
  60000,
);

describe("gumtree.ts", () => {
  test("exports live scraper function", () => {
    expect(getGumtreeValues).toBeTypeOf("function");
  });
});
