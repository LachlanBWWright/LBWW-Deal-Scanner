import mongoose, { Schema } from "mongoose";

export interface globalsInterface {
  BOT_CLIENT_ID: string;
  CASH_CONVERTERS: boolean;
  CASH_CONVERTERS_CHANNEL_ID: string | null;
  CASH_CONVERTERS_ROLE_ID: string | null;
  COMMAND_PERMISSION_ROLE_ID: string | null;
  CS_CHANNEL_ID: string | null;
  CS_ITEMS: boolean;
  CS_MARKET_CHANNEL_ID: string | null;
  CS_MARKET_ROLE_ID: string | null;
  CS_ROLE_ID: string | null;
  DISCORD_GUILD_ID: string;
  DISCORD_TOKEN: string;
  EBAY: boolean;
  EBAY_CHANNEL_ID: string | null;
  EBAY_ROLE_ID: string | null;
  ERROR_CHANNEL_ID: string | null;
  GUMTREE: boolean;
  GUMTREE_CHANNEL_ID: string | null;
  GUMTREE_ROLE_ID: string | null;
  SALVOS: boolean;
  SALVOS_CHANNEL_ID: string | null;
  SALVOS_ROLE_ID: string | null;
  STEAM_QUERY: boolean;
  STEAM_QUERY_CHANNEL_ID: string | null;
  STEAM_QUERY_ROLE_ID: string | null;
}

const GlobalsSchema = new Schema<globalsInterface>({
  BOT_CLIENT_ID: { type: String, required: true },
  CASH_CONVERTERS: { type: Boolean, required: true },
  CASH_CONVERTERS_CHANNEL_ID: { type: String, default: null },
  CASH_CONVERTERS_ROLE_ID: { type: String, default: null },
  COMMAND_PERMISSION_ROLE_ID: { type: String, default: null },
  CS_CHANNEL_ID: { type: String, default: null },
  CS_ITEMS: { type: Boolean, required: true },
  CS_MARKET_CHANNEL_ID: { type: String, default: null },
  CS_MARKET_ROLE_ID: { type: String, default: null },
  CS_ROLE_ID: { type: String, default: null },
  DISCORD_GUILD_ID: { type: String, required: true },
  DISCORD_TOKEN: { type: String, required: true },
  EBAY: { type: Boolean, required: true },
  EBAY_CHANNEL_ID: { type: String, default: null },
  EBAY_ROLE_ID: { type: String, default: null },
  ERROR_CHANNEL_ID: { type: String, default: null },
  GUMTREE: { type: Boolean, required: true },
  GUMTREE_CHANNEL_ID: { type: String, default: null },
  GUMTREE_ROLE_ID: { type: String, default: null },
  SALVOS: { type: Boolean, required: true },
  SALVOS_CHANNEL_ID: { type: String, default: null },
  SALVOS_ROLE_ID: { type: String, default: null },
  STEAM_QUERY: { type: Boolean, required: true },
  STEAM_QUERY_CHANNEL_ID: { type: String, default: null },
  STEAM_QUERY_ROLE_ID: { type: String, default: null },
});

const Globals = mongoose.model<globalsInterface>("Globals", GlobalsSchema);

export default Globals;
