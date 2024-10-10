import puppeteer from "puppeteer";
import { test } from "vitest";
import { getGumtreeValues } from "./gumtree";

test("gumtree scanner", async () => {
  const browser = await puppeteer.launch({
    //headless: "shell",
    args: [
      "--no-sandbox",
      "--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
    ],
  });
  let page = await browser.newPage();
  const results = await getGumtreeValues(page, {
    url: "https://www.gumtree.com.au/s-xbox/ballarat-city/k0c21119l3000626r10",
    maxPrice: 50,
  });
  console.log(results);
});
