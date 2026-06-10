import {
  SteamMarketResultsSchema,
  SteamCsMarketResponseSchema,
  SteamItemInfoResponseSchema,
} from "./apiValidators.js";
import { fetchAndValidateJson } from "./fetchJson.js";

export async function fetchSteamMarketResults(url: string) {
  return fetchAndValidateJson(
    url,
    SteamMarketResultsSchema,
    undefined,
    "Failed to fetch Steam Market results",
  ).map((data) => data.results);
}

export async function fetchSteamCsMarketListing(url: string) {
  return fetchAndValidateJson(
    url,
    SteamCsMarketResponseSchema,
    undefined,
    "Failed to fetch and validate Steam CS Market listing",
  );
}

export async function fetchSteamItemInfo(url: string) {
  return fetchAndValidateJson(
    url,
    SteamItemInfoResponseSchema,
    undefined,
    "Failed to fetch and validate Steam item info",
  );
}
