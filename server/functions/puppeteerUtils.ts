import { Page, ElementHandle, Browser } from "puppeteer";
import { resultAsync } from "./neverthrowUtils.js";

export async function gotoPageSafely(
  page: Page,
  url: string,
  options: {
    waitUntil?: "load" | "domcontentloaded" | "networkidle0" | "networkidle2";
    timeout?: number;
  } = {}
) {
  return resultAsync(
    () =>
      page.goto(url, {
        waitUntil: options.waitUntil ?? "domcontentloaded",
        timeout: options.timeout ?? 10000,
      }),
    `Failed to navigate to ${url}`
  );
}

export async function waitForSelectorSafely(
  page: Page,
  selector: string,
  options: { timeout?: number } = {}
) {
  return resultAsync(
    () =>
      page.waitForSelector(selector, {
        timeout: options.timeout ?? 10000,
      }),
    `Failed to wait for selector: ${selector}`
  );
}

export async function querySelectorSafely(
  element: ElementHandle<Element> | Page,
  selector: string
) {
  return resultAsync(
    () => element.$(selector),
    `Failed to query selector: ${selector}`
  );
}

export async function querySelectorAllSafely(
  element: ElementHandle<Element> | Page,
  selector: string
) {
  return resultAsync(
    () => element.$$(selector),
    `Failed to query all for selector: ${selector}`
  );
}

export async function querySelectorEvalSafely<R>(
  element: ElementHandle<Element> | Page,
  selector: string,
  pageFunction: (element: Element) => R
) {
  return resultAsync(
    () => element.$eval(selector, pageFunction),
    `Failed to eval selector: ${selector}`
  );
}

export async function evaluateSafely<R>(
  element: ElementHandle<Element>,
  pageFunction: (element: Element) => R
) {
  return resultAsync(
    () => element.evaluate(pageFunction),
    "Failed to evaluate element"
  );
}

export async function getPageTitleSafely(page: Page) {
  return resultAsync(
    () => page.title(),
    "Failed to get page title"
  );
}

export async function setUserAgentSafely(page: Page, userAgent: string) {
  return resultAsync(
    () => page.setUserAgent(userAgent),
    "Failed to set user agent"
  );
}

export async function closePageSafely(page: Page) {
  return resultAsync(
    () => page.close(),
    "Failed to close page"
  );
}

export async function createNewPageSafely(browser: Browser) {
  return resultAsync(
    () => browser.newPage(),
    "Failed to create new page"
  );
}
