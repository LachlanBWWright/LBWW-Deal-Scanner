import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { db } from "../../globals/PrismaClient.js";

interface CommandInteraction {
  readonly options: {
    getString(name: string): string | null;
    getBoolean(name: string): boolean | null;
  };
  editReply(message: string): Promise<unknown>;
}

export default async function (interaction: CommandInteraction) {
  const query = interaction.options.getString("query") || "placeholder";
  const dmOnly = interaction.options.getBoolean("dmonly") ?? false;
  const search = new URL(query);
  if (search.toString().includes("https://www.cashconverters.com.au/")) {
    await db.query.create({
      data: {
        dmOnly,
        cashConverters: {
          create: { url: search.toString() },
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
