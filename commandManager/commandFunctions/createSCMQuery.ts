import { ChatInputCommandInteraction } from "discord.js";
import { getCsQueryString } from "../../scanners/siteScanners/steamMarket.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { db } from "../../globals/PrismaClient.js";
import { chromium } from "playwright";

export default async function (interaction: ChatInputCommandInteraction) {
  let query = interaction.options.getString("query") || "placeholder";
  let maxPrice = interaction.options.getNumber("maxprice") || 1;

  const browser = await chromium.launch({
    headless: true,
    args: ["--no-sandbox"],
  });
  let page = await browser.newPage();
  try {
    const newUrl = await getCsQueryString(page, query);

    await db.steamMarket.create({
      data: {
        name: newUrl,
        displayUrl: query,
        maxPrice: maxPrice,
        lastPrice: 0,
      },
    });

    await interaction.editReply(
      `${getResponsePrelude()}, the item was added successfully! URL generated: ${newUrl}`,
    );
  } catch (e) {
    await interaction.editReply(`${getFailurePrelude()} the URL is invalid!`);
  }
  await page.close();
  await browser.close();
}
