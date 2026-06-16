export type DealSource =
  | "cashConverters"
  | "ebay"
  | "gumtree"
  | "salvos"
  | "steamMarket"
  | "csTrade"
  | "lootFarm"
  | "tradeIt";

export type QueryType =
  | "cashConverters"
  | "ebay"
  | "gumtree"
  | "salvos"
  | "steamMarket"
  | "csTradeBot"
  | "csMarket";

export interface DealNotification {
  kind: "deal";
  source: DealSource;
  title: string;
  url: string;
  price?: number;
  imageUrl?: string;
  query?: {
    type: QueryType;
    id: string;
    dmOnly?: boolean;
  };
  tags?: string[];
}

export interface ErrorNotification {
  kind: "error";
  source: string;
  message: string;
  stack?: string;
}

export type AppNotification = DealNotification | ErrorNotification;
