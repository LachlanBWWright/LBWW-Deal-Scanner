import puppeteer from "puppeteer";
import SalvosQuery from "../../schema/salvosQuery.js";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import selectorRace from "../../functions/selectorRace.js";
import sendToChannel from "../../functions/sendToChannel.js";

let cursor = SalvosQuery.find().cursor();

export async function scanSalvos(page: puppeteer.Page) {
  if (!globals.SALVOS || !globals.SALVOS_CHANNEL_ID || !globals.SALVOS_ROLE_ID)
    return;
  setStatus("Scanning Salvos");

  let item = await cursor.next();
  if (item === null) {
    cursor = SalvosQuery.find().cursor();
    item = await cursor.next();
  }

  await page.goto(item.name);

  const selector = await selectorRace(
    page,
    ".line-clamp-3",
    ".py-16.flex.flex-col.items-center.justify-center.text-center.text-gray-500"
  );

  if (!selector) return;

  const price = await page.$eval(".product-price", (item) =>
    item.textContent ? parseFloat(item.textContent.slice(1)) : null
  );
  if (price === null) return;

  const imageURL = await page.$eval(
    `img[class="absolute top-0 left-0 w-full h-full object-cover object-center"]`,
    (image) => image.getAttribute("src")
  );

  let salvosItem = await selector?.evaluate((el) => el.textContent);
  if (salvosItem != item.lastItemFound) {
    sendToChannel(
      globals.SALVOS_CHANNEL_ID,
      `<@&${globals.SALVOS_ROLE_ID}> Please know that a ${salvosItem} is available for $${price} at ${item.name} \n\n${imageURL}`
    );

    if (salvosItem != undefined) {
      item.lastItemFound = salvosItem;
      item.save();
    }
  }
}
