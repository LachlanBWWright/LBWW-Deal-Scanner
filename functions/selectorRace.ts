import { Page, Locator } from "playwright";

//Accepts a playwright page, a selector for items found, and a selector for the "nothing found" text
//Returns the locator if items were found, and null if nothing was found
export default async function selectorRace(
  page: Page,
  foundSelector: string,
  noItemSelector: string,
) {
  return await Promise.race<Locator | null>([
    new Promise((res) => {
      page
        .waitForSelector(foundSelector)
        .then(() => res(page.locator(foundSelector)))
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
