import { SlashCommandBuilder } from "@discordjs/builders";

export default new SlashCommandBuilder()
  .setName("createcsmarket")
  .setDescription("Creates a search for a CS item on the SCM.")
  .addStringOption((option) =>
    option
      .setName("query")
      .setDescription("The URL of the query.")
      .setRequired(true),
  )
  .addNumberOption((option) =>
    option
      .setName("maxfloat")
      .setDescription("Enter the max float value for a notification.")
      .setRequired(true),
  )
  .addNumberOption((option) =>
    option
      .setName("maxprice")
      .setDescription("Enter the maximum price for a notification.")
      .setRequired(true),
  )
  .addBooleanOption((option) =>
    option
      .setName("dmonly")
      .setDescription("Make this query DM-only (no channel notifications)")
      .setRequired(false),
  );

export const editcsmarket = new SlashCommandBuilder()
  .setName("editcsmarket")
  .setDescription("Edit an existing CS Market saved query")
  .addStringOption((option) =>
    option
      .setName("id")
      .setDescription("ID of the saved query to edit")
      .setRequired(true),
  )
  .addStringOption((option) =>
    option
      .setName("query")
      .setDescription("New query string")
      .setRequired(true),
  );

export const deletecsmarket = new SlashCommandBuilder()
  .setName("deletecsmarket")
  .setDescription("Delete an existing CS Market saved query")
  .addStringOption((option) =>
    option
      .setName("id")
      .setDescription("ID of the saved query to delete")
      .setRequired(true),
  );

export const viewcsmarketqueries = new SlashCommandBuilder()
  .setName("viewcsmarketqueries")
  .setDescription("List all saved CS Market queries");
