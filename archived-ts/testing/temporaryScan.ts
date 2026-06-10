import puppeteer, { type Page } from "puppeteer";
import { ResultAsync } from "neverthrow";
import { toError } from "../functions/neverthrowUtils.js";
import type { DealNotification, QueryType } from "../deals/types.js";
import {
  fetchSteamCsMarketListing,
  fetchSteamItemInfo,
  fetchSteamMarketResults,
} from "../functions/steamMarketFetcher.js";
import { fetchCsTradeItems } from "../functions/csTraceFetcher.js";
import type { ManualScanItemResult } from "./types.js";
import type { DealSource } from "../deals/types.js";

export interface TemporaryScanResult {
  items: ManualScanItemResult[];
  notifications: DealNotification[];
  errors: string[];
}

const browserScannerTypes: QueryType[] = [
  "cashConverters",
  "ebay",
  "gumtree",
  "salvos",
];

function optionalNumber(value: unknown, fallback: number) {
  if (value === undefined || value === null || value === "") return fallback;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function requiredNumber(
  value: unknown,
  name: string,
  errors: string[],
): number | null {
  if (value === undefined || value === null || value === "") {
    errors.push(`${name} is required`);
    return null;
  }
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) {
    errors.push(`${name} must be a number`);
    return null;
  }
  return parsed;
}

function phraseList(value: unknown) {
  return String(value ?? "")
    .split(",")
    .map((phrase) => phrase.trim().toLowerCase())
    .filter((phrase) => phrase.length > 0);
}

function matchesPhraseFilters(
  text: string,
  requiredPhrases: unknown,
  excludePhrases: unknown,
) {
  const normalized = text.toLowerCase();
  const required = phraseList(requiredPhrases);
  const excluded = phraseList(excludePhrases);
  return (
    required.every((phrase) => normalized.includes(phrase)) &&
    excluded.every((phrase) => !normalized.includes(phrase))
  );
}

function parseCurrency(value: string | null): number | null {
  if (!value) return null;
  const parsed = Number(value.replace(/[^0-9.-]+/g, ""));
  return Number.isFinite(parsed) ? parsed : null;
}

async function collectEbayItems(
  page: Page,
  url: string,
  maxPrice: number,
): Promise<ManualScanItemResult[]> {
  await page.goto(url, { waitUntil: "domcontentloaded", timeout: 10000 });
  await page.waitForSelector(
    "div.srp-river, .srp-save-null-search__heading",
    { timeout: 10000 },
  );
  const rawItems = await page.$$eval(
    "div.su-card-container.su-card-container--horizontal",
    (cards) =>
      cards.slice(0, 20).map((card) => {
        const title =
          card.querySelector('div[role="heading"]')?.textContent?.trim() ??
          "";
        const priceText =
          card.querySelector("span.s-card__price")?.textContent?.trim() ??
          null;
        const link =
          card.querySelector<HTMLAnchorElement>("a[href]")?.href ?? "";
        const imageUrl =
          card.querySelector<HTMLImageElement>("img")?.src ?? null;
        return { title, priceText, link, imageUrl };
      }),
  );

  return rawItems
    .filter((item) => item.title || item.link)
    .map((item) => {
      const price = parseCurrency(item.priceText);
      const passedFilters = price !== null && price <= maxPrice;
      const title = item.title.replace(/^New listing/i, "").trim();
      return {
        source: "ebay",
        title,
        url: item.link || url,
        price,
        imageUrl: item.imageUrl,
        passedFilters,
        filterReason: filterReason(
          passedFilters,
          price === null
            ? "Price could not be read"
            : `Price ${price} is above max ${maxPrice}`,
        ),
      };
    });
}

async function collectGumtreeItems(
  page: Page,
  url: string,
  maxPrice: number,
): Promise<ManualScanItemResult[]> {
  await page.goto(url, { waitUntil: "domcontentloaded", timeout: 10000 });
  await page.setUserAgent(
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
  );
  await page.waitForSelector("a.user-ad-row-new-design", { timeout: 10000 });
  const rawItems = await page.$$eval("a.user-ad-row-new-design", (cards) =>
    cards.slice(0, 20).map((card) => {
      const title =
        card
          .querySelector(".user-ad-row-new-design__title span")
          ?.textContent?.trim() ?? "";
      const priceText =
        card
          .querySelector(".user-ad-price-new-design__price")
          ?.textContent?.trim() ?? null;
      const link = card.href;
      const imageUrl = card.querySelector<HTMLImageElement>("img")?.src ?? null;
      return { title, priceText, link, imageUrl };
    }),
  );

  return rawItems
    .filter((item) => item.title || item.link)
    .map((item) => {
      const price = item.priceText?.includes("Free")
        ? 0
        : parseCurrency(item.priceText);
      const passedFilters = price !== null && price <= maxPrice;
      return {
        source: "gumtree",
        title: item.title,
        url: item.link || url,
        price,
        imageUrl: item.imageUrl,
        passedFilters,
        filterReason: filterReason(
          passedFilters,
          price === null
            ? "Price could not be read"
            : `Price ${price} is above max ${maxPrice}`,
        ),
      };
    });
}

async function collectCashConvertersItems(
  page: Page,
  url: string,
  requiredPhrases: unknown,
  excludePhrases: unknown,
): Promise<ManualScanItemResult[]> {
  await page.goto(url, { waitUntil: "domcontentloaded", timeout: 10000 });
  await page.waitForSelector(".product-item, .no-search-results__text", {
    timeout: 10000,
  });
  const rawItems = await page.$$eval(".product-item", (cards) =>
    cards.slice(0, 20).map((card) => {
      const title =
        card
          .querySelector(".product-item__title__description")
          ?.textContent?.trim() ?? "";
      const priceText =
        card.querySelector(".product-item__price")?.textContent?.trim() ??
        null;
      const postageText =
        card.querySelector(".product-item__postage")?.textContent?.trim() ??
        null;
      const link =
        card.querySelector<HTMLAnchorElement>("a[href]")?.href ?? "";
      const imageUrl =
        card.querySelector<HTMLImageElement>("img")?.src ?? null;
      return { title, priceText, postageText, link, imageUrl };
    }),
  );

  return rawItems
    .filter((item) => item.title || item.link)
    .map((item) => {
      const price = parseCurrency(item.priceText);
      const postage = parseCurrency(item.postageText);
      const totalPrice =
        price === null ? null : price + (postage === null ? 0 : postage);
      const passedFilters = matchesPhraseFilters(
        item.title,
        requiredPhrases,
        excludePhrases,
      );
      return {
        source: "cashConverters",
        title: item.title,
        url: item.link || url,
        price: totalPrice,
        imageUrl: item.imageUrl,
        passedFilters,
        filterReason: filterReason(passedFilters, "Phrase filters did not match"),
      };
    });
}

async function collectSalvosItems(
  page: Page,
  name: string,
  minPrice: number,
  maxPrice: number,
): Promise<ManualScanItemResult[]> {
  const url = `https://www.salvosstores.com.au/shop?search=${encodeURIComponent(
    name,
  )}&sorting=newestFirst`;
  await page.goto(url, { waitUntil: "domcontentloaded", timeout: 10000 });
  await page.waitForSelector("[class*='product-price']", { timeout: 10000 });
  const rawItems = await page.$$eval(
    "div.flex.flex-col.overflow-hidden.rounded.shadow-card.bg-white.h-auto",
    (cards) =>
      cards.slice(0, 20).map((card) => {
        const linkElement =
          card.querySelector<HTMLAnchorElement>(
            "a.mt-2.text-xs.lg\\:text-base.line-clamp-3",
          ) ?? card.querySelector<HTMLAnchorElement>("a[href]");
        const title = linkElement?.textContent?.trim() ?? "";
        const link = linkElement?.href ?? "";
        const priceText =
          card.querySelector("[class*='product-price']")?.textContent?.trim() ??
          null;
        const imageUrl =
          card.querySelector<HTMLImageElement>("img")?.src ?? null;
        return { title, link, priceText, imageUrl };
      }),
  );

  return rawItems
    .filter((item) => item.title || item.link)
    .map((item) => {
      const price = parseCurrency(item.priceText);
      const passedFilters =
        price !== null && price >= minPrice && price <= maxPrice;
      return {
        source: "salvos",
        title: item.title,
        url: item.link || url,
        price,
        imageUrl: item.imageUrl,
        passedFilters,
        filterReason: filterReason(
          passedFilters,
          price === null
            ? "Price could not be read"
            : `Price ${price} is outside ${minPrice}-${maxPrice}`,
        ),
      };
    });
}

function filterReason(passedFilters: boolean, reason: string) {
  return passedFilters ? null : reason;
}

function notificationFromItem(
  item: ManualScanItemResult,
  type: QueryType,
): DealNotification {
  const source = dealSourceFromManualItem(item.source);
  if (!source) {
    return {
      kind: "deal",
      source: "steamMarket",
      title: item.title,
      url: item.url,
      ...(item.price !== null && { price: item.price }),
      ...(item.imageUrl !== null && { imageUrl: item.imageUrl }),
      query: { type, id: item.url },
    };
  }
  return {
    kind: "deal",
    source,
    title: item.title,
    url: item.url,
    ...(item.price !== null && { price: item.price }),
    ...(item.imageUrl !== null && { imageUrl: item.imageUrl }),
    query: { type, id: item.url },
  };
}

function dealSourceFromManualItem(source: string): DealSource | null {
  switch (source) {
    case "cashConverters":
    case "ebay":
    case "gumtree":
    case "salvos":
    case "steamMarket":
    case "csTrade":
      return source;
    default:
      return null;
  }
}

async function runFetchOnlyTemporaryScan(
  type: QueryType,
  payload: Record<string, unknown>,
): Promise<TemporaryScanResult> {
  const items: ManualScanItemResult[] = [];
  const notifications: DealNotification[] = [];
  const errors: string[] = [];

  switch (type) {
    case "steamMarket": {
      const name = String(payload.name ?? "").trim();
      const displayUrl = String(payload.displayUrl || name);
      const maxPrice = requiredNumber(payload.maxPrice, "maxPrice", errors);
      if (!name) errors.push("steamMarket scan requires payload.name");
      if (!name || maxPrice === null) break;

      const result = await fetchSteamMarketResults(name);
      if (result.isErr()) {
        errors.push(result.error.message);
        break;
      }

      for (const listing of result.value.slice(0, 20)) {
        const price = listing.sell_price / 100.0;
        const passedFilters = price <= maxPrice;
        const item: ManualScanItemResult = {
          source: "steamMarket",
          title: listing.name,
          url: displayUrl,
          price,
          imageUrl: null,
          passedFilters,
          filterReason: filterReason(
            passedFilters,
            `Price ${price} is above max ${maxPrice}`,
          ),
        };
        items.push(item);
        if (passedFilters) notifications.push(notificationFromItem(item, type));
      }
      break;
    }

    case "csMarket": {
      const url = String(payload.url ?? "").trim();
      const displayUrl = String(payload.displayUrl || url);
      const maxPrice = requiredNumber(payload.maxPrice, "maxPrice", errors);
      const maxFloat = requiredNumber(payload.maxFloat, "maxFloat", errors);
      if (!url) errors.push("csMarket scan requires payload.url");
      if (!url || maxPrice === null || maxFloat === null) break;

      const listingResult = await fetchSteamCsMarketListing(url);
      if (listingResult.isErr()) {
        errors.push(listingResult.error.message);
        break;
      }

      let checked = 0;
      for (const listing of Object.values(listingResult.value.listinginfo)) {
        if (checked >= 10) break;
        checked++;
        const action = listing.asset.market_actions[0];
        if (!action) continue;
        const itemUrl = "https://api.csgofloat.com/?url="
          .concat(action.link)
          .replace("%listingid%", listing.listingid)
          .replace("%assetid%", listing.asset.id);
        const price =
          (listing.converted_price_per_unit + listing.converted_fee_per_unit) /
          100.0;
        const itemInfoResult = await fetchSteamItemInfo(itemUrl);
        if (itemInfoResult.isErr()) {
          errors.push(itemInfoResult.error.message);
          continue;
        }
        const itemInfo = itemInfoResult.value.iteminfo;
        const passedFilters = itemInfo.floatvalue <= maxFloat && price <= maxPrice;
        const item: ManualScanItemResult = {
          source: "steamMarket",
          title: `${itemInfo.full_item_name} with float ${itemInfo.floatvalue}`,
          url: displayUrl,
          price,
          imageUrl: null,
          passedFilters,
          filterReason: filterReason(
            passedFilters,
            itemInfo.floatvalue > maxFloat
              ? `Float ${itemInfo.floatvalue} is above max ${maxFloat}`
              : `Price ${price} is above max ${maxPrice}`,
          ),
        };
        items.push(item);
        if (passedFilters) notifications.push(notificationFromItem(item, type));
      }
      break;
    }

    case "csTradeBot": {
      const name = String(payload.name ?? "").trim();
      const maxPrice = requiredNumber(payload.maxPrice, "maxPrice", errors);
      const minFloat = requiredNumber(payload.minFloat, "minFloat", errors);
      const maxFloat = requiredNumber(payload.maxFloat, "maxFloat", errors);
      if (!name) errors.push("csTradeBot scan requires payload.name");
      if (
        !name ||
        maxPrice === null ||
        minFloat === null ||
        maxFloat === null
      ) {
        break;
      }

      const itemsResult = await fetchCsTradeItems();
      if (itemsResult.isErr()) {
        errors.push(itemsResult.error.message);
        break;
      }

      for (const fetchedItem of itemsResult.value) {
        const nameMatches = fetchedItem.market_hash_name === name;
        const passedFilters =
          nameMatches &&
          fetchedItem.price <= maxPrice &&
          fetchedItem.wear >= minFloat &&
          fetchedItem.wear <= maxFloat;
        const item: ManualScanItemResult = {
          source: "csTrade",
          title: `${fetchedItem.market_hash_name} with float ${fetchedItem.wear}`,
          url: "https://cs.trade/",
          price: fetchedItem.price,
          imageUrl: null,
          passedFilters,
          filterReason: filterReason(
            passedFilters,
            !nameMatches
              ? "Name does not match"
              : fetchedItem.price > maxPrice
                ? `Price ${fetchedItem.price} is above max ${maxPrice}`
                : `Float ${fetchedItem.wear} is outside ${minFloat}-${maxFloat}`,
          ),
        };
        items.push(item);
        if (passedFilters) {
          notifications.push({
            kind: "deal",
            source: "csTrade",
            title: item.title,
            url: item.url,
            price: item.price ?? undefined,
            imageUrl: item.imageUrl ?? undefined,
            query: { type, id: name },
          });
        }
      }
      break;
    }

    default:
      errors.push(`Scanner type "${type}" is not supported for temporary scans`);
  }

  return { items, notifications, errors };
}

export async function runTemporaryScan(
  type: QueryType,
  payload: Record<string, unknown>,
): Promise<TemporaryScanResult> {
  const items: ManualScanItemResult[] = [];
  const notifications: DealNotification[] = [];
  const errors: string[] = [];

  if (!browserScannerTypes.includes(type)) {
    return await runFetchOnlyTemporaryScan(type, payload);
  }

  const browserResult = await ResultAsync.fromThrowable(
    () =>
      puppeteer.launch({
        args: ["--no-sandbox"],
        headless: true,
      }),
    (error) => toError(error, "Failed to launch temporary scan browser"),
  )();

  if (browserResult.isErr()) {
    return { items, notifications, errors: [browserResult.error.message] };
  }
  const browser = browserResult.value;

  const pageResult = await ResultAsync.fromThrowable(
    () => browser.newPage(),
    (error) => toError(error, "Failed to create temporary scan page"),
  )();

  if (pageResult.isErr()) {
    await ResultAsync.fromThrowable(
      () => browser.close(),
      (error) => toError(error, "Failed to close browser"),
    )();
    return { items, notifications, errors: [pageResult.error.message] };
  }

  const page = pageResult.value;

  switch (type) {
    case "ebay": {
      const url = String(payload.url ?? "").trim();
      const maxPrice = optionalNumber(payload.maxPrice, Infinity);
      if (!url) {
        errors.push("ebay scan requires payload.url");
        break;
      }

      const scanResult = await ResultAsync.fromThrowable(
        () => collectEbayItems(page, url, maxPrice),
        (error) => toError(error, "eBay scan failed"),
      )();

      if (scanResult.isErr()) {
        errors.push(scanResult.error.message);
        break;
      }

      items.push(...scanResult.value);
      for (const item of scanResult.value) {
        if (item.passedFilters) {
          notifications.push(notificationFromItem(item, type));
        }
      }
      break;
    }

    case "gumtree": {
      const url = String(payload.url ?? "").trim();
      const maxPrice = optionalNumber(payload.maxPrice, Infinity);
      if (!url) {
        errors.push("gumtree scan requires payload.url");
        break;
      }

      const scanResult = await ResultAsync.fromThrowable(
        () => collectGumtreeItems(page, url, maxPrice),
        (error) => toError(error, "Gumtree scan failed"),
      )();

      if (scanResult.isErr()) {
        errors.push(scanResult.error.message);
        break;
      }

      items.push(...scanResult.value);
      for (const item of scanResult.value) {
        if (item.passedFilters) {
          notifications.push(notificationFromItem(item, type));
        }
      }
      break;
    }

    case "cashConverters": {
      const url = String(payload.url ?? "").trim();
      if (!url) {
        errors.push("cashConverters scan requires payload.url");
        break;
      }

      const scanResult = await ResultAsync.fromThrowable(
        () =>
          collectCashConvertersItems(
            page,
            url,
            payload.requiredPhrases,
            payload.excludePhrases,
          ),
        (error) => toError(error, "Cash Converters scan failed"),
      )();

      if (scanResult.isErr()) {
        errors.push(scanResult.error.message);
        break;
      }

      items.push(...scanResult.value);
      for (const item of scanResult.value) {
        if (item.passedFilters) {
          notifications.push(notificationFromItem(item, type));
        }
      }
      break;
    }

    case "salvos": {
      const name = String(payload.name ?? "").trim();
      const minPrice = optionalNumber(payload.minPrice, 0);
      const maxPrice = optionalNumber(payload.maxPrice, Infinity);
      if (!name) {
        errors.push("salvos scan requires payload.name");
        break;
      }

      const scanResult = await ResultAsync.fromThrowable(
        () => collectSalvosItems(page, name, minPrice, maxPrice),
        (error) => toError(error, "Salvos scan failed"),
      )();

      if (scanResult.isErr()) {
        errors.push(scanResult.error.message);
        break;
      }

      items.push(...scanResult.value);
      for (const item of scanResult.value) {
        if (item.passedFilters) {
          notifications.push(notificationFromItem(item, type));
        }
      }
      break;
    }

    default:
      errors.push(
        `Scanner type "${type}" is not supported for temporary scans`,
      );
  }

  const closePageResult = await ResultAsync.fromThrowable(
    () => page.close(),
    (error) => toError(error, "Failed to close page"),
  )();
  if (closePageResult.isErr()) {
    errors.push(closePageResult.error.message);
  }

  const closeBrowserResult = await ResultAsync.fromThrowable(
    () => browser.close(),
    (error) => toError(error, "Failed to close browser"),
  )();
  if (closeBrowserResult.isErr()) {
    errors.push(closeBrowserResult.error.message);
  }

  return { items, notifications, errors };
}
