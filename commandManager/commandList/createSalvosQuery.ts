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
  )
  .addBooleanOption((option) =>
    option
      .setName("dmonly")
      .setDescription("Make this query DM-only (no channel notifications)")
      .setRequired(false),
  );

export const editsalvosquery = new SlashCommandBuilder()
  .setName("editsalvosquery")
  .setDescription("Edit an existing Salvos saved query")
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

export const deletesalvosquery = new SlashCommandBuilder()
  .setName("deletesalvosquery")
  .setDescription("Delete an existing Salvos saved query")
  .addStringOption((option) =>
    option
      .setName("id")
      .setDescription("ID of the saved query to delete")
      .setRequired(true),
  );

export const viewsalvosqueries = new SlashCommandBuilder()
  .setName("viewsalvosqueries")
  .setDescription("List all saved Salvos queries");
