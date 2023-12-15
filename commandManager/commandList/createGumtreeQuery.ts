import { SlashCommandBuilder } from "discord.js";

export default new SlashCommandBuilder()
  .setName("creategumtreequery")
  .setDescription("Creates a search for a Gumtree Query")
  .addStringOption((option) =>
    option
      .setName("query")
      .setDescription("The URL of the query. Sort by newest first.")
      .setRequired(true)
  )
  .addNumberOption((option) =>
    option
      .setName("maxprice")
      .setDescription("Enter the maximum price (in AUD) a notification.")
      .setRequired(true)
  );
