import puppeteer from "puppeteer";
import { getEbayValues } from "./ebay";
import { test } from "vitest";

test("ebay scanner", async () => {
  const browser = await puppeteer.launch({
    headless: "shell",
    args: ["--no-sandbox"],
  });
  let page = await browser.newPage();
  const { foundName, foundPrice, foundImage } = await getEbayValues(page, {
    url: "https://www.ebay.com.au/sch/i.html?_from=R40&_nkw=xbox&_sacat=0&_sop=10&rt=nc&LH_BIN=1",
    maxPrice: 50,
  });
  console.log(foundName, foundPrice, foundImage);
});
