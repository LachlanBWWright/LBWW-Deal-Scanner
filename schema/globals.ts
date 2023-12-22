import mongoose, { Schema } from "mongoose";

export interface globalsInterface {
  BOT_CLIENT_ID: string;
  CASH_CONVERTERS: boolean;
  CASH_CONVERTERS_CHANNEL_ID: number | null;
  CASH_CONVERTERS_ROLE_ID: number | null;
  COMMAND_PERMISSION_ROLE_ID: number | null;
  CS_CHANNEL_ID: number | null;
  CS_ITEMS: boolean;
  CS_MARKET_CHANNEL_ID: number | null;
  CS_MARKET_ROLE_ID: number | null;
  CS_ROLE_ID: number | null;
  DISCORD_GUILD_ID: string;
  DISCORD_TOKEN: string;
  EBAY: boolean;
  EBAY_CHANNEL_ID: number | null;
  EBAY_ROLE_ID: number | null;
  FACEBOOK: boolean;
  FACEBOOK_CHANNEL_ID: number | null;
  FACEBOOK_ROLE_ID: number | null;
  GUMTREE: boolean;
  GUMTREE_CHANNEL_ID: number | null;
  GUMTREE_ROLE_ID: number | null;
  PS5BIGW: boolean;
  PS5_CHANNEL_ID: number | null;
  PS5_ROLE_ID: number | null;
  SALVOS: boolean;
  SALVOS_CHANNEL_ID: number | null;
  SALVOS_ROLE_ID: number | null;
  STEAM_QUERY: boolean;
  STEAM_QUERY_CHANNEL_ID: number | null;
  STEAM_QUERY_ROLE_ID: number | null;
  XBOXBIGW: boolean;
  XBOX_CHANNEL_ID: number | null;
  XBOX_ROLE_ID: number | null;
}

const GlobalsSchema = new Schema<globalsInterface>({
  BOT_CLIENT_ID: { type: String, required: true },
  CASH_CONVERTERS: { type: Boolean, required: true },
  CASH_CONVERTERS_CHANNEL_ID: { type: Number, default: null },
  CASH_CONVERTERS_ROLE_ID: { type: Number, default: null },
  COMMAND_PERMISSION_ROLE_ID: { type: Number, default: null },
  CS_CHANNEL_ID: { type: Number, default: null },
  CS_ITEMS: { type: Boolean, required: true },
  CS_MARKET_CHANNEL_ID: { type: Number, default: null },
  CS_MARKET_ROLE_ID: { type: Number, default: null },
  CS_ROLE_ID: { type: Number, default: null },
  DISCORD_GUILD_ID: { type: String, required: true },
  DISCORD_TOKEN: { type: String, required: true },
  EBAY: { type: Boolean, required: true },
  EBAY_CHANNEL_ID: { type: Number, default: null },
  EBAY_ROLE_ID: { type: Number, default: null },
  FACEBOOK: { type: Boolean, required: true },
  FACEBOOK_CHANNEL_ID: { type: Number, default: null },
  FACEBOOK_ROLE_ID: { type: Number, default: null },
  GUMTREE: { type: Boolean, required: true },
  GUMTREE_CHANNEL_ID: { type: Number, default: null },
  GUMTREE_ROLE_ID: { type: Number, default: null },
  PS5BIGW: { type: Boolean, required: true },
  PS5_CHANNEL_ID: { type: Number, default: null },
  PS5_ROLE_ID: { type: Number, default: null },
  SALVOS: { type: Boolean, required: true },
  SALVOS_CHANNEL_ID: { type: Number, default: null },
  SALVOS_ROLE_ID: { type: Number, default: null },
  STEAM_QUERY: { type: Boolean, required: true },
  STEAM_QUERY_CHANNEL_ID: { type: Number, default: null },
  STEAM_QUERY_ROLE_ID: { type: Number, default: null },
  XBOXBIGW: { type: Boolean, required: true },
  XBOX_CHANNEL_ID: { type: Number, default: null },
  XBOX_ROLE_ID: { type: Number, default: null },
});

const GlobalsQuery = mongoose.model<globalsInterface>(
  "GlobalsQuery",
  GlobalsSchema
);

export default GlobalsQuery;
