import { SlashCommandBuilder } from "discord.js";

new SlashCommandBuilder()
.setName("viewqueries")
.setDescription("View the queries created for a website")
.addStringOption((option) =>
  option
    .setName("whichscanner")
    .setDescription("Which part of the scanner's queries are to be viewed")
    .addChoices(
      { name: "CS Deals", value: "csdeals" },
      { name: "Cs.Trade", value: "cstrade" },
      { name: "TradeIt", value: "tradeit" },
      { name: "Loot.Farm", value: "lootfarm" },
      { name: "SCM Query", value: "scmquery" },
      { name: "CS Market", value: "csmarket" },
      { name: "Cash Converters", value: "cashquery" },
      { name: "Salvos", value: "salvosquery" },
      { name: "Ebay", value: "ebayquery" },
      { name: "Gumtree", value: "gumtreequery" },
    )
    .setRequired(true),
)
.addStringOption((option) =>
  option
    .setName("searchname")
    .setDescription(
      "Find a search by name (URL for query, skin name for CS bot.)",
    )
    .setRequired(false),
)

  new SlashCommandBuilder()
    .setName("viewqueries")
    .setDescription("View the queries created for a website")
    .addStringOption((option) =>
      option
        .setName("whichscanner")
        .setDescription("Which part of the scanner's queries are to be viewed")
        .addChoices(
          { name: "CS Deals", value: "csdeals" },
          { name: "Cs.Trade", value: "cstrade" },
          { name: "TradeIt", value: "tradeit" },
          { name: "Loot.Farm", value: "lootfarm" },
          { name: "SCM Query", value: "scmquery" },
          { name: "CS Market", value: "csmarket" },
          { name: "Cash Converters", value: "cashquery" },
          { name: "Salvos", value: "salvosquery" },
          { name: "Ebay", value: "ebayquery" },
          { name: "Gumtree", value: "gumtreequery" },
        )
        .setRequired(true),
    )
    .addStringOption((option) =>
      option
        .setName("searchname")
        .setDescription(
          "Find a search by name (URL for query, skin name for CS bot.)",
        )
        .setRequired(false),
    )