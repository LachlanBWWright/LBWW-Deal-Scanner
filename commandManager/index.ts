import { CacheType, Interaction, GuildMember, Client } from "discord.js";

//Definitions for slash command parameters
import createCashQueryDefinition from "./commandList/createCashQuery.js";
import createCSMarketDefinition from "./commandList/createCSMarket.js";
import createEbayQueryDefinition from "./commandList/createEbayQuery.js";
import createGumtreeQueryDefinition from "./commandList/createGumtreeQuery.js";
import createMultiSearchDefinition from "./commandList/createMultiSearchQuery.js";
import createSalvosQueryDefinition from "./commandList/createSalvosQuery.js";
import createSCMQueryDefinition from "./commandList/createSCMQuery.js";
import deleteQueryQueryDefinition from "./commandList/deleteQueryQuery.js";
import viewQueriesQueryDefinition from "./commandList/viewQueriesQuery.js";

//Functions that run after a slash command is sent
import createCashQuery from "./commandFunctions/commands/createCashQuery.js";
import createCSMarket from "./commandFunctions/commands/createCSMarket.js";
import createEbayQuery from "./commandFunctions/commands/createEbayQuery.js";
import createGumtreeQuery from "./commandFunctions/commands/createGumtreeQuery.js";
import createMultiSearch from "./commandFunctions/commands/createMultiSearchQuery.js";
import createSalvosQuery from "./commandFunctions/commands/createSalvosQuery.js";
import createSCMQuery from "./commandFunctions/commands/createSCMQuery.js";
import deleteQueryQuery from "./commandFunctions/commands/deleteQueryQuery.js";
import viewQueriesQuery from "./commandFunctions/commands/viewQueriesQuery.js";

//Command handler code
export const commandList = [
  createCashQueryDefinition,
  createCSMarketDefinition,
  createEbayQueryDefinition,
  createGumtreeQueryDefinition,
  createMultiSearchDefinition,
  createSalvosQueryDefinition,
  createSCMQueryDefinition,
  deleteQueryQueryDefinition,
  viewQueriesQueryDefinition,
];

export async function commandHandler(
  interaction: Interaction<CacheType>,
  client: Client
) {
  if (!interaction.isChatInputCommand()) return; //Cancels if not a command
  try {
    await interaction.deferReply(); //Creates the loading '...'

    let roleFound = false;
    let member = interaction.member;
    member = <GuildMember>member;
    member.roles.cache.map((role) => {
      if (role.id == process.env.COMMAND_PERMISSION_ROLE_ID) {
        roleFound = true;
      }
    });
    if (!roleFound) {
      await interaction.editReply(
        "Please know you don't have the role needed to make commands. That's not acceptable."
      );
      return;
    }

    if (interaction.commandName === "createcashquery")
      await createCashQuery(interaction);
    else if (interaction.commandName === "createcsmarket")
      await createCSMarket(client, interaction);
    else if (interaction.commandName === "createebayquery")
      await createEbayQuery(interaction);
    else if (interaction.commandName === "creategumtreequery")
      await createGumtreeQuery(interaction);
    else if (interaction.commandName === "createmultisearch")
      await createMultiSearch(client, interaction);
    else if (interaction.commandName === "createsalvosquery")
      await createSalvosQuery(interaction);
    else if (interaction.commandName === "createscmquery")
      await createSCMQuery(client, interaction);
    else if (interaction.commandName === "deletequery")
      await deleteQueryQuery(interaction);
    else if (interaction.commandName === "viewqueries")
      await viewQueriesQuery(interaction);
  } catch (err) {
    console.error(err);
    await interaction.editReply(
      `Please know that an error has occurred: ${err}`
    );
  }
}
