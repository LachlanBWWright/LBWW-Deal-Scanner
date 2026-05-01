import { ChatInputCommandInteraction } from "discord.js";
import { getCsQueryString } from "../../scanners/siteScanners/steamMarket.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { db } from "../../globals/PrismaClient.js";
import puppeteer from "puppeteer";
import { fromThrowableAsync } from "../../functions/neverthrowUtils.js";

export default async function (interaction: ChatInputCommandInteraction) {
  const query = interaction.options.getString("query") || "placeholder";
  const maxPrice = interaction.options.getNumber("maxprice") || 1;
  const dmOnly = interaction.options.getBoolean("dmonly") ?? false;

  const browserResult = await fromThrowableAsync(
    () =>
      puppeteer.launch({
        headless: "shell",
        args: ["--no-sandbox"],
      }),
    "Failed to launch browser",
  );
  if (browserResult.isErr()) {
    await interaction.editReply(`${getFailurePrelude()} the URL is invalid!`);
    return;
  }

  const browser = browserResult.value;
  const pageResult = await fromThrowableAsync(
    () => browser.newPage(),
    "Failed to create browser page",
  );
  if (pageResult.isErr()) {
    await fromThrowableAsync(() => browser.close(), "Failed to close browser");
    await interaction.editReply(`${getFailurePrelude()} the URL is invalid!`);
    return;
  }

  const page = pageResult.value;
  const createResult = await fromThrowableAsync(async () => {
    const newUrl = await getCsQueryString(page, query);
    if (!newUrl) {
      throw new Error("Invalid URL");
    }

    await db.query.create({
      data: {
        dmOnly,
        steamMarket: {
          create: {
            name: newUrl,
            displayUrl: query,
            maxPrice,
            lastPrice: 0,
          },
        },
      },
    });

    return newUrl;
  }, "Failed to create SCM query");

  await fromThrowableAsync(() => page.close(), "Failed to close page");
  await fromThrowableAsync(() => browser.close(), "Failed to close browser");

  if (createResult.isErr()) {
    await interaction.editReply(`${getFailurePrelude()} the URL is invalid!`);
    return;
  }

  await interaction.editReply(
    `${getResponsePrelude()}, the item was added successfully${dmOnly ? " (DM only)" : ""}! URL generated: ${createResult.value}`,
  );
}
