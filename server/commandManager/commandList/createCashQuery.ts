import { SlashCommandBuilder } from "@discordjs/builders";

export default new SlashCommandBuilder()
  .setName("createcashquery")
  .setDescription("Creates a search for a Cash Converters Query")
  .addStringOption((option) =>
    option
      .setName("query")
      .setDescription("The URL of the query. Sort by price or newness.")
      .setRequired(true),
  )
  .addBooleanOption((option) =>
    option
      .setName("dmonly")
      .setDescription("Make this query DM-only (no channel notifications)")
      .setRequired(false),
  );

// Additional commands (edit / delete / view) for Cash Converters queries
export const editcashquery = new SlashCommandBuilder()
  .setName("editcashquery")
  .setDescription("Edit an existing Cash Converters saved query")
  .addStringOption((option) =>
    option
      .setName("id")
      .setDescription("ID of the saved query to edit")
      .setRequired(true),
  )
  .addStringOption((option) =>
    option
      .setName("query")
      .setDescription("New URL or query string to replace the old one")
      .setRequired(true),
  );

export const deletecashquery = new SlashCommandBuilder()
  .setName("deletecashquery")
  .setDescription("Delete an existing Cash Converters saved query")
  .addStringOption((option) =>
    option
      .setName("id")
      .setDescription("ID of the saved query to delete")
      .setRequired(true),
  );

export const viewcashqueries = new SlashCommandBuilder()
  .setName("viewcashqueries")
  .setDescription("List all saved Cash Converters queries");
