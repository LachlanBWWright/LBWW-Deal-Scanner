import puppeteer from "puppeteer";
import { getCashConvertersValues } from "./cashConverters.js";
import { describe, expect, test } from "vitest";

test("cash converters scanner returns latest listing shape", async () => {
  const browser = await puppeteer.launch({
    headless: "shell",
    args: ["--no-sandbox"],
  });

  const page = await browser.newPage();
  const result = await getCashConvertersValues(page, {
    url: "https://www.cashconverters.com.au/search-results?Sort=default&page=1&f%5Bcategory%5D%5B0%5D=all&f%5Blocations%5D%5B0%5D=all&query=xbox",
  });

  expect(result).toBeDefined();
  expect(result?.itemName).toBeTypeOf("string");
  expect(result?.totalPrice).toBeTypeOf("number");
  expect((result?.totalPrice ?? -1) >= 0).toBe(true);
  expect(result?.image).toMatch(/^https?:\/\//);

  await browser.close();
}, 60000);

describe("cashConverters.ts", () => {
  test("exports live scraper function", () => {
    expect(getCashConvertersValues).toBeTypeOf("function");
  });
});
