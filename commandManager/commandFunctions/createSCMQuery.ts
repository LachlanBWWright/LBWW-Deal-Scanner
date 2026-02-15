import { ChatInputCommandInteraction } from "discord.js";
import { getCsQueryString } from "../../scanners/siteScanners/steamMarket.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { db } from "../../globals/PrismaClient.js";
import puppeteer from "puppeteer";

export default async function (interaction: ChatInputCommandInteraction) {
  const query = interaction.options.getString("query") || "placeholder";
  const maxPrice = interaction.options.getNumber("maxprice") || 1;
  const dmOnly = interaction.options.getBoolean("dmonly") ?? false;

  const browser = await puppeteer.launch({
    headless: "shell",
    args: ["--no-sandbox"],
  });
  const page = await browser.newPage();
  try {
    const newUrl = await getCsQueryString(page, query);

    await db.query.create({
      data: {
        dmOnly,
        steamMarket: {
          create: {
            name: newUrl,
            displayUrl: query,
            maxPrice: maxPrice,
            lastPrice: 0,
          },
        },
      },
    });

    if (status)
      await interaction.editReply(
        `${getResponsePrelude()}, the item was added successfully${dmOnly ? " (DM only)" : ""}! URL generated: ${newUrl}`,
      );
    else
      await interaction.editReply(`${getFailurePrelude()} the URL is invalid!`);
  } catch {
    await interaction.editReply(`${getFailurePrelude()} the URL is invalid!`);
  }
  await page.close();
  await browser.close();
}
