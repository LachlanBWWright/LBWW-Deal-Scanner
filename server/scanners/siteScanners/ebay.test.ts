import puppeteer from "puppeteer";
import { getEbayValues } from "./ebay.js";
import { test } from "vitest";

test("ebay scanner", async () => {
  const browser = await puppeteer.launch({
    args: ["--no-sandbox"],
    headless: "shell",
  });
  const page = await browser.newPage();
  await page.setUserAgent(
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
  );
  const { foundName, foundPrice, foundImage } = await getEbayValues(page, {
    url: "https://www.ebay.com.au/sch/i.html?_nkw=ps4&_sacat=0&_from=R40&_sop=10",
    maxPrice: 50,
  });
  console.log(foundName, foundPrice, foundImage);
});
