import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import { db } from "../../globals/PrismaClient.js";
import type { Page } from "puppeteer";
import { resultAsync } from "../../functions/neverthrowUtils.js";
import type { DealNotification } from "../../deals/types.js";
import {
  evaluateListingForQuery,
  matchPriceRange,
  persistListingObservation,
  type DiscoveredListing,
} from "../listingState.js";

interface SalvosQueryInput {
  name: string;
  minPrice: number;
  maxPrice: number;
}

interface SalvosValues {
  readonly name: string;
  readonly price: number;
  readonly image: string | null;
  readonly link: string;
}

const salvosGridSelector =
  "div[class='grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5 lg:gap-10']";
const salvosCardSelector =
  "div[class='flex flex-col overflow-hidden rounded shadow-card bg-white h-auto']";

export async function scanSalvos(page: Page): Promise<DealNotification[]> {
  if (!globals.SALVOS) return [];
  setStatus("Scanning Salvos");

  const queries = await db.salvos.findMany();
  const notifications: DealNotification[] = [];

  for (const item of queries) {
    const queryId = await ensureSalvosQueryId(item.name, item.queryId);
    const foundItems = await getSalvosListings(page, item);

    for (const foundItem of foundItems) {
      const listing = normalizeSalvosListing(foundItem);
      const persistedResult = await persistListingObservation({ listing });
      if (persistedResult.isErr()) {
        console.warn(persistedResult.error.message);
        continue;
      }

      const decisionResult = await evaluateListingForQuery({
        queryId,
        listingId: persistedResult.value.listingId,
        source: "salvos",
        totalPrice: listing.totalPrice,
        match: matchPriceRange(listing.totalPrice, item.minPrice, item.maxPrice),
      });
      if (decisionResult.isErr()) {
        console.warn(decisionResult.error.message);
        continue;
      }

      if (decisionResult.value.type === "Notify") {
        notifications.push({
          kind: "deal",
          source: "salvos",
          title: `a ${listing.title} is available for $${listing.totalPrice ?? "unknown"} at ${listing.canonicalUrl}`,
          url: listing.canonicalUrl,
          price: listing.totalPrice ?? undefined,
          imageUrl: listing.imageUrl ?? undefined,
          query: {
            type: "salvos",
            id: item.name,
          },
        });
      }
    }
  }

  return notifications;
}

export async function getSalvosListings(
  page: Page,
  item: SalvosQueryInput,
): Promise<SalvosValues[]> {
  const gotoResult = await resultAsync(
    () =>
      page.goto(
        `https://www.salvosstores.com.au/shop?search=${encodeURIComponent(
          item.name,
        )}&sorting=newestFirst&price=0-99999`,
        { waitUntil: "domcontentloaded", timeout: 10000 },
      ),
    "Salvos navigation failed",
  );
  if (gotoResult.isErr()) {
    console.warn(gotoResult.error.message);
    return [];
  }

  const extractionResult = await resultAsync(
    () =>
      page.$$eval(
        `${salvosGridSelector} ${salvosCardSelector}`,
        (cards): SalvosValues[] =>
          cards
            .map((card): SalvosValues | null => {
              const parsePrice = (input: string | null): number | null => {
                if (!input) return null;
                const value = Number.parseFloat(input.replace(/[^0-9.]+/g, ""));
                return Number.isFinite(value) ? value : null;
              };
              const linkElement = card.querySelector(
                "a[class='mt-2 text-xs lg:text-base line-clamp-3'], a[href]",
              );
              const name = linkElement?.textContent?.trim() ?? null;
              const link =
                linkElement instanceof HTMLAnchorElement
                  ? linkElement.href
                  : null;
              const priceText =
                card
                  .querySelector(
                    "div[class='font-medium lg:font-semibold text-xs lg:text-xl product-price'], [class*='product-price']",
                  )
                  ?.textContent?.trim() ?? null;
              const imageElement = card.querySelector("img");
              const image =
                imageElement instanceof HTMLImageElement ? imageElement.src : null;
              const price = parsePrice(priceText);

              if (!name || !link || price === null) return null;
              return { name, price, image, link };
            })
            .filter((card): card is SalvosValues => card !== null),
      ),
    "Salvos card extraction failed",
  );
  if (extractionResult.isErr()) {
    console.warn(extractionResult.error.message);
    return [];
  }

  return dedupeByUrl(extractionResult.value);
}

export async function getSalvosValues(page: Page, item: SalvosQueryInput) {
  const listings = await getSalvosListings(page, item);
  return listings[0];
}

function normalizeSalvosListing(item: SalvosValues): DiscoveredListing {
  return {
    source: "salvos",
    externalId: null,
    canonicalUrl: item.link,
    title: item.name,
    price: item.price,
    shipping: null,
    totalPrice: item.price,
    currency: "AUD",
    imageUrl: item.image,
    description: null,
    availability: "available",
  };
}

async function ensureSalvosQueryId(
  name: string,
  queryId: string | null,
): Promise<string> {
  if (queryId) return queryId;
  const query = await db.query.create({
    data: {
      salvos: {
        connect: { name },
      },
    },
  });
  return query.id;
}

function dedupeByUrl(items: SalvosValues[]): SalvosValues[] {
  const seen = new Set<string>();
  const results: SalvosValues[] = [];
  for (const item of items) {
    if (seen.has(item.link)) continue;
    seen.add(item.link);
    results.push(item);
  }
  return results;
}

/* 
Depreciated - Return to scraping
function getUrlCode(e: string) {
  const t = "Product";
  if (!e) return "";
  const n = D(e);
  const r = /(\w+):([\da-f-]+)/.exec(n);
  if (!r) return null;
  if (t && t !== r[1]) return null;
  return r[2];
}

function A(e: string) {
  return Buffer.from(e, "base64").toString("utf8");
}
function D(e: string) {
  return A(M(e));
}
function M(e: string) {
  return m(e.replace(/[-_]/g, (e) => ("-" == e ? "+" : "/")));
}
function m(e: string) {
  return e.replace(/[^A-Za-z0-9\+\/]/g, "");
}
 */
