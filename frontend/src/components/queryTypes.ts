export type QueryType =
  | "cashConverters"
  | "ebay"
  | "gumtree"
  | "salvos"
  | "csMarket"
  | "steamMarket"
  | "csTradeBot";

export interface QueryItem {
  type: QueryType;
  id: string;
  dmOnly: boolean;
  url?: string;
  name?: string;
  displayUrl?: string;
  maxPrice?: number;
  minPrice?: number;
  minFloat?: number;
  maxFloat?: number;
  requiredPhrases?: string;
  excludePhrases?: string;
}

export interface QueryFormState {
  url: string;
  name: string;
  displayUrl: string;
  maxPrice: number;
  minPrice: number;
  minFloat: number;
  maxFloat: number;
  requiredPhrases: string;
  excludePhrases: string;
  dmOnly: boolean;
}

export type FormField = {
  label: string;
  key: keyof QueryFormState;
  type: string;
};

export const supportedQueryTypes: QueryType[] = [
  "cashConverters",
  "ebay",
  "gumtree",
  "salvos",
  "csMarket",
  "steamMarket",
  "csTradeBot",
];
