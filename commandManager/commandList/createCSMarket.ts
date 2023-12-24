import { SlashCommandBuilder } from "@discordjs/builders";

export default new SlashCommandBuilder()
  .setName("createcsmarket")
  .setDescription("Creates a search for a CS item on the SCM.")
  .addStringOption((option) =>
    option
      .setName("query")
      .setDescription("The URL of the query.")
      .setRequired(true)
  )
  .addNumberOption((option) =>
    option
      .setName("maxfloat")
      .setDescription("Enter the max float value for a notification.")
      .setRequired(true)
  )
  .addNumberOption((option) =>
    option
      .setName("maxprice")
      .setDescription("Enter the maximum price for a notification.")
      .setRequired(true)
  );
