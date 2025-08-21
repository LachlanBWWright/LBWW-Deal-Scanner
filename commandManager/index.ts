import { CacheType, Interaction, GuildMember } from "discord.js";

//Definitions for slash command parameters
import createCashQueryDefinition, {
  editcashquery,
  deletecashquery,
  viewcashqueries,
} from "./commandList/createCashQuery.js";
import createCSMarketDefinition, {
  editcsmarket,
  deletecsmarket,
  viewcsmarketqueries,
} from "./commandList/createCSMarket.js";
import createEbayQueryDefinition, {
  editedbayquery,
  deleteebayquery,
  viewebayqueries,
} from "./commandList/createEbayQuery.js";
import createGumtreeQueryDefinition, {
  editgumtreequery,
  deletegumtreequery,
  viewgumtreequeries,
} from "./commandList/createGumtreeQuery.js";
import createMultiSearchDefinition, {
  editmultisearchquery,
  deletemultisearchquery,
  viewmultisearchqueries,
} from "./commandList/createMultiSearchQuery.js";
import createSalvosQueryDefinition, {
  editsalvosquery,
  deletesalvosquery,
  viewsalvosqueries,
} from "./commandList/createSalvosQuery.js";
import createSCMQueryDefinition, {
  editscmquery,
  deletescmquery,
  viewscmqueries,
} from "./commandList/createSCMQuery.js";
/* import deleteQueryQueryDefinition from "./commandList/deleteQueryQuery.js";
import viewQueriesQueryDefinition from "./commandList/viewQueriesQuery.js"; */

//Functions that run after a slash command is sent
import createCashQuery from "./commandFunctions/createCashQuery.js";
import createCSMarket from "./commandFunctions/createCSMarket.js";
import createEbayQuery from "./commandFunctions/createEbayQuery.js";
import createGumtreeQuery from "./commandFunctions/createGumtreeQuery.js";
import createMultiSearch from "./commandFunctions/createMultiSearchQuery.js";
import createSalvosQuery from "./commandFunctions/createSalvosQuery.js";
import createSCMQuery from "./commandFunctions/createSCMQuery.js";
/* import deleteQueryQuery from "./commandFunctions/deleteQueryQuery.js";
import viewQueriesQuery from "./commandFunctions/viewQueriesQuery.js"; */
import editCashQuery from "./commandFunctions/editCashQuery.js";
import deleteCashQuery from "./commandFunctions/deleteCashQuery.js";
import viewCashQueries from "./commandFunctions/viewCashQueries.js";

import editEbayQuery from "./commandFunctions/editEbayQuery.js";
import deleteEbayQuery from "./commandFunctions/deleteEbayQuery.js";
import viewEbayQueries from "./commandFunctions/viewEbayQueries.js";

import editGumtreeQuery from "./commandFunctions/editGumtreeQuery.js";
import deleteGumtreeQuery from "./commandFunctions/deleteGumtreeQuery.js";
import viewGumtreeQueries from "./commandFunctions/viewGumtreeQueries.js";

import editCSMarket from "./commandFunctions/editCSMarket.js";
import deleteCSMarket from "./commandFunctions/deleteCSMarket.js";
import viewCSMarketQueries from "./commandFunctions/viewCSMarketQueries.js";

import editMultiSearchQuery from "./commandFunctions/editMultiSearchQuery.js";
import deleteMultiSearchQuery from "./commandFunctions/deleteMultiSearchQuery.js";
import viewMultiSearchQueries from "./commandFunctions/viewMultiSearchQueries.js";

import editSCMQuery from "./commandFunctions/editSCMQuery.js";
import deleteSCMQuery from "./commandFunctions/deleteSCMQuery.js";
import viewSCMQueries from "./commandFunctions/viewSCMQueries.js";

import editSalvosQuery from "./commandFunctions/editSalvosQuery.js";
import deleteSalvosQuery from "./commandFunctions/deleteSalvosQuery.js";
import viewSalvosQueries from "./commandFunctions/viewSalvosQueries.js";
import { getFailurePrelude } from "../functions/messagePreludes.js";

//Command handler code
export const commandList = [
  createCashQueryDefinition,
  createCSMarketDefinition,
  createEbayQueryDefinition,
  createGumtreeQueryDefinition,
  createMultiSearchDefinition,
  createSalvosQueryDefinition,
  createSCMQueryDefinition,
  // edit / delete / view definitions
  editcashquery,
  deletecashquery,
  viewcashqueries,
  editcsmarket,
  deletecsmarket,
  viewcsmarketqueries,
  editedbayquery,
  deleteebayquery,
  viewebayqueries,
  editgumtreequery,
  deletegumtreequery,
  viewgumtreequeries,
  editmultisearchquery,
  deletemultisearchquery,
  viewmultisearchqueries,
  editsalvosquery,
  deletesalvosquery,
  viewsalvosqueries,
  editscmquery,
  deletescmquery,
  viewscmqueries,
  /*   deleteQueryQueryDefinition,
  viewQueriesQueryDefinition, */
];

export async function commandHandler(interaction: Interaction<CacheType>) {
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
        `${getFailurePrelude()} you don't have the role needed to make commands.`,
      );
      return;
    }

    if (interaction.commandName === "createcashquery")
      await createCashQuery(interaction);
    else if (interaction.commandName === "createcsmarket")
      await createCSMarket(interaction);
    else if (interaction.commandName === "createebayquery")
      await createEbayQuery(interaction);
    else if (interaction.commandName === "creategumtreequery")
      await createGumtreeQuery(interaction);
    else if (interaction.commandName === "createmultisearch")
      await createMultiSearch(interaction);
    else if (interaction.commandName === "createsalvosquery")
      await createSalvosQuery(interaction);
    else if (interaction.commandName === "createscmquery")
      await createSCMQuery(interaction);
    // Edit / Delete / View handlers
    else if (interaction.commandName === "editcashquery")
      await editCashQuery(interaction);
    else if (interaction.commandName === "deletecashquery")
      await deleteCashQuery(interaction);
    else if (interaction.commandName === "viewcashqueries")
      await viewCashQueries(interaction);
    else if (interaction.commandName === "editcsmarket")
      await editCSMarket(interaction);
    else if (interaction.commandName === "deletecsmarket")
      await deleteCSMarket(interaction);
    else if (interaction.commandName === "viewcsmarketqueries")
      await viewCSMarketQueries(interaction);
    else if (interaction.commandName === "editedbayquery")
      await editEbayQuery(interaction);
    else if (interaction.commandName === "deleteebayquery")
      await deleteEbayQuery(interaction);
    else if (interaction.commandName === "viewebayqueries")
      await viewEbayQueries(interaction);
    else if (interaction.commandName === "editgumtreequery")
      await editGumtreeQuery(interaction);
    else if (interaction.commandName === "deletegumtreequery")
      await deleteGumtreeQuery(interaction);
    else if (interaction.commandName === "viewgumtreequeries")
      await viewGumtreeQueries(interaction);
    else if (interaction.commandName === "editmultisearchquery")
      await editMultiSearchQuery(interaction);
    else if (interaction.commandName === "deletemultisearchquery")
      await deleteMultiSearchQuery(interaction);
    else if (interaction.commandName === "viewmultisearchqueries")
      await viewMultiSearchQueries(interaction);
    else if (interaction.commandName === "editsalvosquery")
      await editSalvosQuery(interaction);
    else if (interaction.commandName === "deletesalvosquery")
      await deleteSalvosQuery(interaction);
    else if (interaction.commandName === "viewsalvosqueries")
      await viewSalvosQueries(interaction);
    else if (interaction.commandName === "editscmquery")
      await editSCMQuery(interaction);
    else if (interaction.commandName === "deletescmquery")
      await deleteSCMQuery(interaction);
    else if (interaction.commandName === "viewscmqueries")
      await viewSCMQueries(interaction);
    /*     else if (interaction.commandName === "deletequery")
      await deleteQueryQuery(interaction);
    else if (interaction.commandName === "viewqueries")
      await viewQueriesQuery(interaction); */
  } catch (err) {
    console.error(err);
    await interaction.editReply(
      `${getFailurePrelude()} an error has occurred: ${err}`,
    );
  }
}
