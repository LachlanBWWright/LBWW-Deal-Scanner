import { Page, ElementHandle } from "puppeteer";

//Accepts a puppeteer page, a selector for items found, and a selector for the "nothing found" text
//Returns the page if items were found, and null if nothing was found
export default async function selectorRace(
  page: Page,
  foundSelector: string,
  noItemSelector: string,
  timeoutMs = 10000,
) {
  return await Promise.race<ElementHandle<Element> | null>([
    page.waitForSelector(foundSelector, { timeout: timeoutMs }).catch(() => null),
    page
      .waitForSelector(noItemSelector, { timeout: timeoutMs })
      .then(() => null)
      .catch(() => null),
  ]);
}
