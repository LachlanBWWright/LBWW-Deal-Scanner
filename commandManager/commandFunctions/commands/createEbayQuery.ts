import { ChatInputCommandInteraction } from "discord.js";
import EbayQuery from "../../../schema/ebayQuery.js";

export default async function (interaction: ChatInputCommandInteraction) {
  let query = interaction.options.getString("query") || "placeholder";
  let maxPrice = interaction.options.getNumber("maxprice") || 1000;
  let search = new URL(query);
  if (search.toString().includes("https://www.ebay.com.au/")) {
    let ebayQuery = new EbayQuery({
      name: search.toString(),
      maxPrice: maxPrice,
    });
    ebayQuery.save();
    await interaction.editReply(
      "Please know that the search has been created: " + search.toString()
    );
  } else interaction.editReply("Please know that your search was invalid!");
}
