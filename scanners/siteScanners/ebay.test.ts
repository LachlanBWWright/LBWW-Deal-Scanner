import puppeteer from "puppeteer";
import { getEbayValues } from "./ebay";
import { test } from "vitest";

test("ebay scanner", async () => {
  const browser = await puppeteer.launch({
    //args: ["--no-sandbox"],
    headless: true,
  });
  let page = await browser.newPage();
  //page.setUserAgent(
  //  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
  //);
  const { foundName, foundPrice, foundImage } = await getEbayValues(page, {
    url: "https://www.ebay.com.au/sch/i.html?_nkw=ps4&_sacat=0&_from=R40&_trksid=p4432023.m570.l1313",
    maxPrice: 50,
  });
  console.log(foundName, foundPrice, foundImage);
});
