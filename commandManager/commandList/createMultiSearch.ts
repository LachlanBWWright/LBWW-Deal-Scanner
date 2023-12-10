
import {SlashCommandBuilder} from "discord.js"

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
        .setDescription(
          "Enter the minimum acceptable float value for the skin.",
        )
        .setRequired(true),
    )
    .addNumberOption((option) =>
      option
        .setName("maxfloat")
        .setDescription(
          "Enter the maximum acceptable float value for the skin.",
        )
        .setRequired(true),
    )
    .addNumberOption((option) =>
      option
        .setName("maxpricecsdeals")
        .setDescription(
          "Enter the max price for the skin on cs.deals, no search will be created if it's $0.",
        )
        .setRequired(false),
    )
    .addNumberOption((option) =>
      option
        .setName("maxpricecstrade")
        .setDescription(
          "Enter the max price for the skin on cs.trade, no search will be created if it's $0.",
        )
        .setRequired(false),
    )
    .addNumberOption((option) =>
      option
        .setName("maxpricetradeit")
        .setDescription(
          "Enter the max price for the skin on tradeit.gg, no search will be created if it's $0.",
        )
        .setRequired(false),
    )
    .addNumberOption((option) =>
      option
        .setName("maxpricelootfarm")
        .setDescription(
          "Enter the max price for the skin on loot.farm, no search will be created if it's $0.",
        )
        .setRequired(false),
    )