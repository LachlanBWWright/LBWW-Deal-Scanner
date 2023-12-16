import createCashQuery from "./commandList/createCashQuery";
import createCSMarket from "./commandList/createCSMarket";
import createEbayQuery from "./commandList/createEbayQuery";
import createGumtreeQuery from "./commandList/createGumtreeQuery";
import createMultiSearch from "./commandList/createMultiSearchQuery";
import createSalvosQuery from "./commandList/createSalvosQuery";
import createSCMQuery from "./commandList/createSCMQuery";
import deleteQueryQuery from "./commandList/deleteQueryQuery";
import viewQueriesQuery from "./commandList/viewQueriesQuery";

//Command handler code
export const commandList = [
  createCashQuery,
  createCSMarket,
  createEbayQuery,
  createGumtreeQuery,
  createMultiSearch,
  createSalvosQuery,
  createSCMQuery,
  deleteQueryQuery,
  viewQueriesQuery,
];

export async function commandHandler() {
  console.log("TEST");
}
