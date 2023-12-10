import { SlashCommandBuilder } from "discord.js";

export default new SlashCommandBuilder()
.setName("createebayquery")
.setDescription("Creates a search for an eBay Query")
.addStringOption((option) =>
  option
    .setName("query")
    .setDescription("The URL of the query. Sort by newest first.")
    .setRequired(true),
)
.addNumberOption((option) =>
  option
    .setName("maxprice")
    .setDescription("Enter the maximum price (in AUD) a notification.")
    .setRequired(true),
)