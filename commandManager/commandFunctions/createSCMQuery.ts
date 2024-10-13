import { ChatInputCommandInteraction } from "discord.js";
import { getCsQueryString } from "../../scanners/siteScanners/steamMarket.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { db } from "../../globals/PrismaClient.js";
import puppeteer from "puppeteer";

export default async function (interaction: ChatInputCommandInteraction) {
  let query = interaction.options.getString("query") || "placeholder";
  let maxPrice = interaction.options.getNumber("maxprice") || 1;

  const browser = await puppeteer.launch({
    headless: "shell",
    args: ["--no-sandbox"],
  });
  let page = await browser.newPage();
  try {
    const newUrl = await getCsQueryString(page, query);

    db.steamMarket.create({
      data: {
        name: newUrl,
        displayUrl: query,
        maxPrice: maxPrice,
        lastPrice: 0,
      },
    });

    if (status)
      await interaction.editReply(
        `${getResponsePrelude()}, the item was added successfully! URL generated: ${newUrl}`,
      );
    else
      await interaction.editReply(`${getFailurePrelude()} the URL is invalid!`);
  } catch (e) {
    await interaction.editReply(`${getFailurePrelude()} the URL is invalid!`);
  }
  await page.close();
  await browser.close();
}
