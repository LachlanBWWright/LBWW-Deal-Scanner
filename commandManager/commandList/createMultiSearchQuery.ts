import { SlashCommandBuilder } from "discord.js";

export default new SlashCommandBuilder()
  .setName("createmultisearch")
  .setDescription("Creates a search for a CS:GO item on multiple trade bots")
  .addStringOption((option) =>
    option
      .setName("skinname")
      .setDescription(
        'Copy and paste from the SCM, E.G. "StatTrak™ XM1014 | Seasons (Minimal Wear)".',
      )
      .setRequired(true),
  )
  .addNumberOption((option) =>
    option
      .setName("minfloat")
      .setDescription("Enter the minimum acceptable float value for the skin.")
      .setRequired(true),
  )
  .addNumberOption((option) =>
    option
      .setName("maxfloat")
      .setDescription("Enter the maximum acceptable float value for the skin.")
      .setRequired(true),
  )
  .addNumberOption((option) =>
    option
      .setName("maxprice")
      .setDescription(
        "Enter the max price for the skin, relative to USD on the Steam Community Market.",
      )
      .setRequired(false),
  );
