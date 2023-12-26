import { TextChannel } from "discord.js";
import puppeteer from "puppeteer";
import SalvosQuery from "../../schema/salvosQuery.js";
import globals from "../../globals/Globals.js";
import client from "../../globals/DiscordJSClient.js";
import setStatus from "../../functions/setStatus.js";
import selectorRace from "../../functions/selectorRace.js";

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

  let salvosItem = await selector?.evaluate((el) => el.textContent);
  if (salvosItem != item.lastItemFound) {
    client.channels
      .fetch(globals.SALVOS_CHANNEL_ID)
      .then((channel) => <TextChannel>channel)
      .then((channel) => {
        if (channel)
          channel.send(
            `<@&${globals.SALVOS_ROLE_ID}> Please know that a ${salvosItem} is available at  ${item.name}`
          );
      })
      .catch((e) => console.error(e));
    if (salvosItem != undefined) {
      item.lastItemFound = salvosItem;
      item.save();
    }
  }
}
