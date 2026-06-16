import puppeteer from "puppeteer";
import { describe, expect, test } from "vitest";
import { getSalvosValues } from "./salvos.js";

const runLiveApiTests = process.env.RUN_LIVE_API_TESTS === "true";

test.skipIf(!runLiveApiTests)(
  "salvos scanner returns first matching listing shape",
  async () => {
  const browser = await puppeteer.launch({
    headless: "shell",
    args: ["--no-sandbox"],
  });

  const page = await browser.newPage();
  const result = await getSalvosValues(page, {
    name: "shirt",
    minPrice: 0,
    maxPrice: 50,
  });

  expect(result === undefined || typeof result === "object").toBe(true);
  if (result) {
    expect(result.name).toBeTypeOf("string");
    expect(result.price).toBeTypeOf("number");
    expect(result.image).toMatch(/^https?:\/\//);
    expect(result.link).toMatch(/^https?:\/\//);
  }

  await browser.close();
  },
  60000,
);

describe("salvos.ts", () => {
  test("exports live scraper function", () => {
    expect(getSalvosValues).toBeTypeOf("function");
  });
});
