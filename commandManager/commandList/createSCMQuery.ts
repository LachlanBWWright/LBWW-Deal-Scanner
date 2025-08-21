import { SlashCommandBuilder } from "discord.js";

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
  );

export const editscmquery = new SlashCommandBuilder()
  .setName("editscmquery")
  .setDescription("Edit an existing SCM saved query")
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

export const deletescmquery = new SlashCommandBuilder()
  .setName("deletescmquery")
  .setDescription("Delete an existing SCM saved query")
  .addStringOption((option) =>
    option
      .setName("id")
      .setDescription("ID of the saved query to delete")
      .setRequired(true),
  );

export const viewscmqueries = new SlashCommandBuilder()
  .setName("viewscmqueries")
  .setDescription("List all saved SCM queries");
