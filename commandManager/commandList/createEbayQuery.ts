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
  .addBooleanOption((option) =>
    option
      .setName("dmonly")
      .setDescription("Make this query DM-only (no channel notifications)")
      .setRequired(false),
  );

export const editedbayquery = new SlashCommandBuilder()
  .setName("editedbayquery")
  .setDescription("Edit an existing eBay saved query")
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

export const deleteebayquery = new SlashCommandBuilder()
  .setName("deleteebayquery")
  .setDescription("Delete an existing eBay saved query")
  .addStringOption((option) =>
    option
      .setName("id")
      .setDescription("ID of the saved query to delete")
      .setRequired(true),
  );

export const viewebayqueries = new SlashCommandBuilder()
  .setName("viewebayqueries")
  .setDescription("List all saved eBay queries");
