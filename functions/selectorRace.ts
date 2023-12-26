import puppeteer from "puppeteer";

//Accepts a puppeteer page, a selector for items found, and a selector for the "nothing found" text
//Returns the page if items were found, and null if nothing was found
export default async function selectorRace(
  page: puppeteer.Page,
  foundSelector: string,
  noItemSelector: string
) {
  return await Promise.race<puppeteer.ElementHandle<Element> | null>([
    new Promise((res) => {
      page
        .waitForSelector(foundSelector)
        .then((selector) => res(selector))
        .catch(() => res(null));
    }),
    new Promise((res) => {
      page
        .waitForSelector(noItemSelector)
        .then(() => res(null))
        .catch(() => res(null));
    }),
  ]);
}
