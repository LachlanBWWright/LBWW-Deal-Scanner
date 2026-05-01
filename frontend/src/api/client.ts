import createClient from "openapi-fetch";

import type { paths } from "./schema.d.ts";

const secretStorageKey = "dealscannerApiSecret";

export const apiClient = createClient<paths>({
  baseUrl: "",
});

export function getApiSecret() {
  return localStorage.getItem(secretStorageKey) ?? "";
}

export function getApiHeaders() {
  const secret = getApiSecret();
  return secret ? { "x-api-secret": secret } : {};
}

export type ApiStatus =
  paths["/api/status"]["get"]["responses"][200]["content"]["application/json"];
export type ApiCommandList =
  paths["/api/commands"]["get"]["responses"][200]["content"]["application/json"];
export type ApiRunScanResponse =
  paths["/api/scans/run"]["post"]["responses"][200]["content"]["application/json"];
