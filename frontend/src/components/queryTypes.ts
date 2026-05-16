import type { paths } from "../api/schema";

export type QueryItem =
  paths["/api/queries"]["get"]["responses"][200]["content"]["application/json"]["queries"][number];

export type QueryType =
  paths["/api/queries"]["post"]["requestBody"]["content"]["application/json"]["type"];

export interface QueryFormState {
  url: string;
  name: string;
  displayUrl: string;
  maxPrice: number | "";
  minPrice: number | "";
  minFloat: number | "";
  maxFloat: number | "";
  requiredPhrases: string;
  excludePhrases: string;
  dmOnly: boolean;
}

export interface SearchResultItem {
  source: string;
  title: string;
  url: string;
  price: number | null;
  imageUrl: string | null;
  queryType: QueryType | null;
  queryId: string | null;
  foundAt: string;
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
