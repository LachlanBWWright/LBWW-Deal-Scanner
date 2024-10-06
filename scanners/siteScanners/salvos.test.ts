import puppeteer from "puppeteer";
import { test } from "vitest";
import { getSalvosValues } from "./salvos";

test("salvos scanner", async () => {
  const browser = await puppeteer.launch({
    headless: "shell",
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
