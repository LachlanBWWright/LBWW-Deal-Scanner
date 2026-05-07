import createClient from "openapi-fetch";

import type { paths } from "./schema.d.ts";

const secretStorageKey = "dealscannerApiSecret";
const hostStorageKey = "dealscannerApiHost";

export const apiClient = createClient<paths>({
  baseUrl: "",
});

export function getApiSecret() {
  return localStorage.getItem(secretStorageKey) ?? "";
}

export function setApiSecret(secret: string) {
  localStorage.setItem(secretStorageKey, secret);
}

export function getApiHost() {
  return localStorage.getItem(hostStorageKey) ?? "";
}

export function setApiHost(host: string) {
  localStorage.setItem(hostStorageKey, host);
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
export type ApiTestingCapabilities =
  paths["/api/testing/capabilities"]["get"]["responses"][200]["content"]["application/json"];
export type ApiTestingNotificationResult =
  paths["/api/testing/notifications"]["post"]["responses"][200]["content"]["application/json"];
export type ApiTestingNotificationHistory =
  paths["/api/testing/notifications"]["get"]["responses"][200]["content"]["application/json"];
export type ApiTestingScanResult =
  paths["/api/testing/scans/run"]["post"]["responses"][200]["content"]["application/json"];
export type ApiTestingScanHistory =
  paths["/api/testing/runs"]["get"]["responses"][200]["content"]["application/json"];
