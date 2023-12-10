import { SlashCommandBuilder } from "@discordjs/builders";

new SlashCommandBuilder()
.setName("createcashquery")
.setDescription("Creates a search for a Cash Converters Query")
.addStringOption((option) =>
  option
    .setName("query")
    .setDescription("The URL of the query. Sort by price or newness.")
    .setRequired(true),
)