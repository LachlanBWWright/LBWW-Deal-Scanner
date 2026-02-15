import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import sendToChannel from "../../functions/sendToChannel.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";
import { db, SCANNER } from "../../globals/PrismaClient.js";
import { checkIfNew } from "../../functions/handleItemUpdate.js";
import { Salvos } from "@prisma/client";
import { Page } from "puppeteer";

export async function scanSalvos(page: Page) {
  if (!globals.SALVOS || !globals.SALVOS_CHANNEL_ID || !globals.SALVOS_ROLE_ID)
    return;
  setStatus("Scanning Salvos");

  const item = await getSalvosQuery();
  if (!item) return;
  let notificationSent = false;

  const result = await getSalvosValues(page, item);
  if (!result) return;

  if (result.price > item.maxPrice || result.price < item.minPrice) return;

  if ((await checkIfNew(result.image, SCANNER.SALVOS)) && !notificationSent) {
    await sendToChannel(
      globals.SALVOS_CHANNEL_ID,
      `<@&${globals.SALVOS_ROLE_ID}> ${getNotificationPrelude()} a ${
        result.name
      } is available for $${result.price} at ${result.link}`,
      { 
        files: [result.image],
        queryId: item.name, // Use name as the unique identifier for salvos
        queryType: 'salvos'
      },
    );
    notificationSent = true;
  }
}

export async function getSalvosValues(page: Page, item: Salvos) {
  try {
    await page.goto(
      `https://www.salvosstores.com.au/shop?search=${encodeURIComponent(
        item.name,
      )}&sorting=newestFirst&price=${item.minPrice ?? 0}-${
        item.maxPrice ?? 99999
      }`,
      { waitUntil: "domcontentloaded", timeout: 10000 },
    );
  } catch (error) {
    console.warn("Salvos navigation failed:", error);
    return;
  }

  const grid = await page.$(
    "div[class='grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5 lg:gap-10']",
  );
  const pageItem = await grid?.$(
    "div[class='flex flex-col overflow-hidden rounded shadow-card bg-white h-auto']",
  );

  if (!pageItem) return;

  const linkElement = await pageItem.$(
    "a[class='mt-2 text-xs lg:text-base line-clamp-3']",
  );
  const name = await linkElement?.evaluate((el) => el.textContent);
  const link = await linkElement?.evaluate((el) => el.href);

  const priceElement =
    (await pageItem.$(
      "div[class='font-medium lg:font-semibold text-xs lg:text-xl product-price']",
    )) ?? (await pageItem.$("[class*='product-price']"));
  const priceText = await priceElement?.evaluate((el) => el.textContent ?? "");
  const price = priceText ? parseFloat(priceText.replace(/[^0-9.]/g, "")) : null;

  const image = await pageItem.$eval("img", (img) => img.src);

  if (!name || !price || !image || !link) return;
  return { name, price, image, link };
}

let index = 0;
async function getSalvosQuery() {
  const query = await db.salvos.findFirst({
    skip: index++,
  });
  if (query) {
    return query;
  }
  index = 1; //Will find the first query in the line below
  return await db.salvos.findFirst();
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
