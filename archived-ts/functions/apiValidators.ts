import { z } from "zod";

const optionalNumber = z.preprocess((value) => {
  if (value === null || value === undefined || value === "") {
    return undefined;
  }

  const numericValue = Number(value);
  return Number.isNaN(numericValue) ? undefined : numericValue;
}, z.coerce.number().optional());

// LootFarm validators
const LootFarmItemSchema = z.object({
  id: z.coerce.string(),
  f: z.string().optional(), // float value as "5642:100"
  l: z.string().optional(),
  tr: z.coerce.number(),
  td: z.coerce.number().optional(),
  st: z.unknown().optional(),
});

const LootFarmSkinSchema = z.object({
  n: z.string(), // name
  cl: z.string(),
  g: z.coerce.number(),
  t: z.record(z.string()).optional(),
  e: z.string().optional(),
  u: z.record(z.array(LootFarmItemSchema)), // botNumber -> items
  pg: optionalNumber,
  p: optionalNumber, // price in cents
});

export const LootFarmResponseSchema = z.object({
  result: z.record(LootFarmSkinSchema),
});

export type LootFarmResponse = z.infer<typeof LootFarmResponseSchema>;
export type LootFarmSkin = z.infer<typeof LootFarmSkinSchema>;

// CS.Trade validators
const CsTradeItemSchema = z
  .object({
    id: z.coerce.string(),
    app_id: z.coerce.number(),
    market_hash_name: z.string(),
    price: z.coerce.number(),
    wear: z.coerce.number(),
    // allow other unknown props from the API
    icon: z.string().optional(),
    status: z.string().optional(),
    type: z.string().optional(),
    bot: z.string().optional(),
    bot_id: z.string().optional(),
  })
  .passthrough();

export const CsTradeResponseSchema = z.object({
  inventory: z.array(CsTradeItemSchema),
});

export type CsTradeResponse = z.infer<typeof CsTradeResponseSchema>;
export type CsTradeItem = z.infer<typeof CsTradeItemSchema>;

// TradeIt validators
const TradeItItemSchema = z
  .object({
    id: z.coerce.string(),
    price: z.coerce.number(),
    name: z.string(),
    floatValue: optionalNumber,
    floatValues: z.array(z.coerce.number()).nullish(),
    assetId: z.coerce.string().optional(),
    classId: z.coerce.string().optional(),
    steamId: z.coerce.string().optional(),
    gameId: z.string().optional(),
    steamInspectLink: z.string().optional(),
  })
  .passthrough();

export const TradeItResponseSchema = z.object({
  items: z.array(TradeItItemSchema),
});

export type TradeItResponse = z.infer<typeof TradeItResponseSchema>;
export type TradeItItem = z.infer<typeof TradeItItemSchema>;

// Steam Market validators
const SteamMarketItemSchema = z.object({
  name: z.string(),
  sell_price: z.coerce.number(),
  // allow other unknown props
  hash_name: z.string().optional(),
  app_id: z.coerce.number().optional(),
  asset_description: z.unknown().optional(),
});

export const SteamMarketResultsSchema = z.object({
  results: z.array(SteamMarketItemSchema),
});

export type SteamMarketResults = z.infer<typeof SteamMarketResultsSchema>;
export type SteamMarketItem = z.infer<typeof SteamMarketItemSchema>;

// Steam CS Market validators
const SteamListingSchema = z.object({
  listingid: z.coerce.string(),
  converted_price_per_unit: z.coerce.number(),
  converted_fee_per_unit: z.coerce.number(),
  asset: z.object({
    id: z.coerce.string(),
    market_actions: z.array(
      z.object({
        link: z.string(),
      }),
    ),
  }),
});

const SteamItemInfoSchema = z.object({
  floatvalue: z.coerce.number(),
  full_item_name: z.string(),
});

export const SteamCsMarketResponseSchema = z.object({
  listinginfo: z.preprocess(
    (value) => (Array.isArray(value) ? {} : value),
    z.record(SteamListingSchema),
  ),
});

export const SteamItemInfoResponseSchema = z.object({
  iteminfo: SteamItemInfoSchema,
});

export type SteamCsMarketResponse = z.infer<typeof SteamCsMarketResponseSchema>;
export type SteamItemInfoResponse = z.infer<typeof SteamItemInfoResponseSchema>;
