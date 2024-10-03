import { SlashCommandBuilder } from "discord.js";

export default new SlashCommandBuilder()
  .setName("createsalvosquery")
  .setDescription("Creates a search for a Salvos Query")
  .addStringOption((option) =>
    option
      .setName("query")
      .setDescription("The name of the query.")
      .setRequired(true),
  )
  .addNumberOption((option) =>
    option
      .setName("minprice")
      .setDescription("Enter the minimum price for a notification.")
      .setRequired(false),
  )
  .addNumberOption((option) =>
    option
      .setName("maxprice")
      .setDescription("Enter the maximum price for a notification.")
      .setRequired(false),
  );
