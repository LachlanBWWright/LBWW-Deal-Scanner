import { ChatInputCommandInteraction } from "discord.js";
import CashConvertersQuery from "../../../schema/cashConvertersQuery.js";

export default async function (interaction: ChatInputCommandInteraction) {
  let query = interaction.options.getString("query") || "placeholder";
  let search = new URL(query);
  if (search.toString().includes("https://www.cashconverters.com.au/")) {
    let cashQuery = new CashConvertersQuery({
      name: search.toString(),
    });
    cashQuery.save();
    await interaction.editReply(
      "Please know that the search has been created: " + search.toString()
    );
  } else interaction.editReply("Please know that your search was invalid!");
}
