import puppeteer from "puppeteer";
import { ResultAsync } from "neverthrow";
import { toError } from "../functions/neverthrowUtils.js";
import type { DealNotification, QueryType } from "../deals/types.js";
import { getEbayValues } from "../scanners/siteScanners/ebay.js";
import { getGumtreeValues } from "../scanners/siteScanners/gumtree.js";
import { getCashConvertersValues } from "../scanners/siteScanners/cashConverters.js";
import { getSalvosValues } from "../scanners/siteScanners/salvos.js";
import { formatPrice } from "../functions/formatPrice.js";

export interface TemporaryScanResult {
  notifications: DealNotification[];
  errors: string[];
}

export async function runTemporaryScan(
  type: QueryType,
  payload: Record<string, unknown>,
): Promise<TemporaryScanResult> {
  const notifications: DealNotification[] = [];
  const errors: string[] = [];

  const browserResult = await ResultAsync.fromThrowable(
    () =>
      puppeteer.launch({
        args: ["--no-sandbox"],
        headless: true,
      }),
    (error) => toError(error, "Failed to launch temporary scan browser"),
  )();

  if (browserResult.isErr()) {
    return { notifications, errors: [browserResult.error.message] };
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
    return { notifications, errors: [pageResult.error.message] };
  }

  const page = pageResult.value;

  switch (type) {
    case "ebay": {
      const url = String(payload.url ?? "").trim();
      const maxPrice = Number(payload.maxPrice ?? Infinity);
      if (!url) {
        errors.push("ebay scan requires payload.url");
        break;
      }

      const scanResult = await ResultAsync.fromThrowable(
        () => getEbayValues(page, { url }),
        (error) => toError(error, "eBay scan failed"),
      )();

      if (scanResult.isErr()) {
        errors.push(scanResult.error.message);
        break;
      }

      const { foundName, foundPrice, foundImage } = scanResult.value;
      if (foundName && foundPrice !== null && foundPrice <= maxPrice) {
        notifications.push({
          kind: "deal",
          source: "ebay",
          title: `a ${foundName} priced at $${foundPrice} is available at ${url}`,
          url,
          price: foundPrice,
          imageUrl: foundImage ?? undefined,
        });
      }
      break;
    }

    case "gumtree": {
      const url = String(payload.url ?? "").trim();
      const maxPrice = Number(payload.maxPrice ?? Infinity);
      if (!url) {
        errors.push("gumtree scan requires payload.url");
        break;
      }

      const scanResult = await ResultAsync.fromThrowable(
        () => getGumtreeValues(page, { url }),
        (error) => toError(error, "Gumtree scan failed"),
      )();

      if (scanResult.isErr()) {
        errors.push(scanResult.error.message);
        break;
      }

      if (scanResult.value) {
        const { foundName, foundPrice, foundImg } = scanResult.value;
        if (foundPrice <= maxPrice) {
          notifications.push({
            kind: "deal",
            source: "gumtree",
            title: `a ${foundName} priced at $${foundPrice} is available at ${url}`,
            url,
            price: foundPrice,
            ...(foundImg && { imageUrl: foundImg }),
          });
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
        () => getCashConvertersValues(page, { url }),
        (error) => toError(error, "Cash Converters scan failed"),
      )();

      if (scanResult.isErr()) {
        errors.push(scanResult.error.message);
        break;
      }

      if (scanResult.value) {
        const { itemName, totalPrice, image } = scanResult.value;
        notifications.push({
          kind: "deal",
          source: "cashConverters",
          title: `a ${itemName} for ${formatPrice(totalPrice)} is available at ${url}`,
          url,
          price: totalPrice,
          imageUrl: image ?? undefined,
        });
      }
      break;
    }

    case "salvos": {
      const name = String(payload.name ?? "").trim();
      const minPrice = Number(payload.minPrice ?? 0);
      const maxPrice = Number(payload.maxPrice ?? Infinity);
      if (!name) {
        errors.push("salvos scan requires payload.name");
        break;
      }

      const scanResult = await ResultAsync.fromThrowable(
        () => getSalvosValues(page, { name, minPrice, maxPrice }),
        (error) => toError(error, "Salvos scan failed"),
      )();

      if (scanResult.isErr()) {
        errors.push(scanResult.error.message);
        break;
      }

      if (scanResult.value) {
        const { name: itemName, price, image, link } = scanResult.value;
        if (price >= minPrice && price <= maxPrice) {
          notifications.push({
            kind: "deal",
            source: "salvos",
            title: `a ${itemName} is available for $${price} at ${link}`,
            url: link,
            price,
            imageUrl: image,
          });
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

  return { notifications, errors };
}
