import puppeteer from "puppeteer";
import { fromThrowableAsync } from "../functions/neverthrowUtils.js";
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

  const browserResult = await fromThrowableAsync(
    () =>
      puppeteer.launch({
        args: ["--no-sandbox"],
        headless: true,
      }),
    "Failed to launch temporary scan browser",
  );

  if (browserResult.isErr()) {
    return { notifications, errors: [browserResult.error.message] };
  }
  const browser = browserResult.value;

  const pageResult = await fromThrowableAsync(
    () => browser.newPage(),
    "Failed to create temporary scan page",
  );

  if (pageResult.isErr()) {
    await fromThrowableAsync(() => browser.close(), "Failed to close browser");
    return { notifications, errors: [pageResult.error.message] };
  }

  const page = pageResult.value;

  try {
    switch (type) {
      case "ebay": {
        const url = String(payload.url ?? "").trim();
        const maxPrice = Number(payload.maxPrice ?? Infinity);
        if (!url) {
          errors.push("ebay scan requires payload.url");
          break;
        }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const fakeEbayRow = { url, maxPrice, queryId: null } as any;
        const scanResult = await fromThrowableAsync(
          () => getEbayValues(page, fakeEbayRow),
          "eBay scan failed",
        );
        if (scanResult.isErr()) {
          errors.push(scanResult.error.message);
        } else {
          const { foundName, foundPrice, foundImage } = scanResult.value;
          if (foundName && foundPrice !== null && foundPrice <= maxPrice) {
            notifications.push({
              kind: "deal",
              source: "ebay",
              title: `a ${foundName} priced at $${foundPrice} is available at ${url}`,
              url,
              price: foundPrice ?? undefined,
              imageUrl: foundImage ?? undefined,
            });
          }
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
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const fakeGumtreeRow = { url, maxPrice, queryId: null } as any;
        const scanResult = await fromThrowableAsync(
          () => getGumtreeValues(page, fakeGumtreeRow),
          "Gumtree scan failed",
        );
        if (scanResult.isErr()) {
          errors.push(scanResult.error.message);
        } else if (scanResult.value) {
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
        const requiredPhrases = String(payload.requiredPhrases ?? "");
        const excludePhrases = String(payload.excludePhrases ?? "");
        if (!url) {
          errors.push("cashConverters scan requires payload.url");
          break;
        }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const fakeCCRow = {
          url,
          requiredPhrases,
          excludePhrases,
          queryId: null,
        } as any;
        const scanResult = await fromThrowableAsync(
          () => getCashConvertersValues(page, fakeCCRow),
          "Cash Converters scan failed",
        );
        if (scanResult.isErr()) {
          errors.push(scanResult.error.message);
        } else if (scanResult.value) {
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
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const fakeSalvosRow = {
          name,
          minPrice,
          maxPrice,
          queryId: null,
        } as any;
        const scanResult = await fromThrowableAsync(
          () => getSalvosValues(page, fakeSalvosRow),
          "Salvos scan failed",
        );
        if (scanResult.isErr()) {
          errors.push(scanResult.error.message);
        } else if (scanResult.value) {
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
  } finally {
    await fromThrowableAsync(() => page.close(), "Failed to close page");
    await fromThrowableAsync(() => browser.close(), "Failed to close browser");
  }

  return { notifications, errors };
}
