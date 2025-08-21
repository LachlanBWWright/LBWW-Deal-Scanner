import { SlashCommandBuilder } from "discord.js";

export default new SlashCommandBuilder()
  .setName("creategumtreequery")
  .setDescription("Creates a search for a Gumtree Query")
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
  );

export const editgumtreequery = new SlashCommandBuilder()
  .setName("editgumtreequery")
  .setDescription("Edit an existing Gumtree saved query")
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

export const deletegumtreequery = new SlashCommandBuilder()
  .setName("deletegumtreequery")
  .setDescription("Delete an existing Gumtree saved query")
  .addStringOption((option) =>
    option
      .setName("id")
      .setDescription("ID of the saved query to delete")
      .setRequired(true),
  );

export const viewgumtreequeries = new SlashCommandBuilder()
  .setName("viewgumtreequeries")
  .setDescription("List all saved Gumtree queries");
