import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

interface CommandInteraction {
  readonly options: {
    getString(name: string): string | null;
    getNumber(name: string): number | null;
    getBoolean(name: string): boolean | null;
  };
  editReply(message: string): Promise<unknown>;
}

export default async function (interaction: CommandInteraction) {
  const query = interaction.options.getString("query") || "placeholder";
  const maxPrice = interaction.options.getNumber("maxprice") || 1000;
  const dmOnly = interaction.options.getBoolean("dmonly") ?? false;
  const search = new URL(query);
  if (search.toString().includes("https://www.ebay.com.au/")) {
    await db.query.create({
      data: {
        dmOnly,
        ebay: {
          create: {
            url: search.toString(),
            maxPrice,
          },
        },
      },
    });
    await interaction.editReply(
      `${getResponsePrelude()} the search has been created${dmOnly ? " (DM only)" : ""}: ${search.toString()}`,
    );
  } else
    await interaction.editReply(
      `${getFailurePrelude()} your search was invalid!`,
    );
}
