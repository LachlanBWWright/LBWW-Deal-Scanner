import puppeteer from "puppeteer";
import { getCashConvertersValues } from "./cashConverters";
import { test } from "vitest";

test("cash converters scanner", async () => {
  const browser = await puppeteer.launch({
    headless: "shell",
    args: ["--no-sandbox"],
  });
  let page = await browser.newPage();
  const result = await getCashConvertersValues(page, {
    url: "https://www.cashconverters.com.au/search-results?Sort=default&page=1&f%5Bcategory%5D%5B0%5D=all&f%5Blocations%5D%5B0%5D=all&query=xbox",
    requiredPhrases: "",
    excludePhrases: "",
    //maxPrice: 50,
  });
  console.log(result);
});
