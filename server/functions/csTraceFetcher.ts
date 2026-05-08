import { okAsync } from "neverthrow";

import { CsTradeResponseSchema } from "./apiValidators.js";
import { fetchAndValidateJson } from "./fetchJson.js";

const CS_TRADE_URL =
  "https://cdn.cs.trade:8443/api/getInventory?order_by=price_desc&bot=all&_=1651756783463";

export async function fetchCsTradeItems() {
  return fetchAndValidateJson(
    CS_TRADE_URL,
    CsTradeResponseSchema,
    undefined,
    "Failed to fetch and validate CS.Trade items",
  ).andThen((data) => okAsync(data.inventory.filter((item) => item.app_id === 730)));
}

export async function warmupCsTradeCache() {
  return fetchAndValidateJson(
    CS_TRADE_URL,
    CsTradeResponseSchema,
    undefined,
    "Failed to warmup CS.Trade cache",
  ).andThen(() => okAsync(true));
}
