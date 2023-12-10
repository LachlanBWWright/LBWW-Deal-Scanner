import {SlashCommandBuilder} from "discord.js"

export default new SlashCommandBuilder()
.setName("createscmquery")
.setDescription("Creates a search for a SCM Query")
.addStringOption((option) =>
  option
    .setName("query")
    .setDescription("The URL of the query. Sort by ascending price.")
    .setRequired(true),
)
.addNumberOption((option) =>
  option
    .setName("maxprice")
    .setDescription("Enter the maximum price for a notification.")
    .setRequired(true),
)