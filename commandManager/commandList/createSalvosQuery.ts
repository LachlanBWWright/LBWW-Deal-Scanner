import { SlashCommandBuilder } from "discord.js";

export default new SlashCommandBuilder()
.setName("createsalvosquery")
.setDescription("Creates a search for a Salvos Query")
.addStringOption((option) =>
  option
    .setName("query")
    .setDescription("The URL of the query. Sort by newest first.")
    .setRequired(true),
)