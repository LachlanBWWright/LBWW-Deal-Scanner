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
  let notificationSent = false;

  const result = await getSalvosValues(page, item);
  if (!result) return;

  if (result.price > item.maxPrice || result.price < item.minPrice) return;

  if ((await checkIfNew(result.image, SCANNER.SALVOS)) && !notificationSent) {
    sendToChannel(
      globals.SALVOS_CHANNEL_ID,
      `<@&${globals.SALVOS_ROLE_ID}> ${getNotificationPrelude()} a ${
        result.name
      } is available for $${result.price} at ${result.link}`,
      { files: [result.image] },
    );
    notificationSent = true;
  }
}

export async function getSalvosValues(page: Page, item: Salvos) {
  await page.goto(
    `https://www.salvosstores.com.au/shop?search=${encodeURIComponent(
      item.name,
    )}&sorting=newestFirst&price=${item.minPrice ?? 0}-${
      item.maxPrice ?? 99999
    }`,
  );

  const grid = await page.$(
    "div[class='grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5 lg:gap-10']",
  );
  const pageItem = await grid?.$(
    "div[class='flex flex-col overflow-hidden rounded shadow-card bg-white h-auto']",
  );

  if (!pageItem) return;

  const name = await pageItem.$eval(
    "a[class='mt-2 text-xs lg:text-base line-clamp-3']",
    (el) => el.textContent,
  );
  const link = await pageItem.$eval(
    "a[class='mt-2 text-xs lg:text-base line-clamp-3']",
    (el) => el.href,
  );

  const price = await pageItem.$eval(
    "div[class='font-medium lg:font-semibold text-xs lg:text-xl product-price']",
    (el) => {
      return el.textContent ? parseFloat(el.textContent.substring(1)) : null; //Remove leading $1
    },
  );

  const image = await pageItem.$eval("img", (img) => img.src);

  if (!name || !price || !image || !link) return;
  return { name, price, image, link };
}

let index = 0;
async function getSalvosQuery() {
  let query = await db.salvos.findFirst({
    skip: index++,
  });
  if (query) {
    index++;
    return query;
  }
  index = 1; //Will find the first query in the line below
  return await db.salvos.findFirstOrThrow();
}

/* 
Depreciated - Return to scraping
function getUrlCode(e: string) {
  const t = "Product";
  if (!e) return "";
  const n = D(e);
  const r = /(\w+):([\da-f-]+)/.exec(n);
  if (!r) return null;
  if (t && t !== r[1]) throw new Error("Schema is not correct");
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
