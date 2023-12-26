import { ChatInputCommandInteraction } from "discord.js";
import SalvosQuery from "../../schema/salvosQuery.js";

export default async function (interaction: ChatInputCommandInteraction) {
  let query = interaction.options.getString("query") || "placeholder";
  let search = new URL(query);
  if (
    search.toString().includes("https://www.salvosstores.com.au/shop?search=")
  ) {
    let salvosQuery = new SalvosQuery({
      name: search.toString(),
    });
    salvosQuery.save();
    await interaction.editReply(
      "Please know that the search has been created: " + search.toString()
    );
  } else interaction.editReply("Please know that your search was invalid!");
}