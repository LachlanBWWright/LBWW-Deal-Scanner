import { ref } from "vue";
import {
  apiClient,
  getApiHeaders,
  getApiHost,
  type ApiTestingCapabilities,
  type ApiTestingNotificationResult,
  type ApiTestingScanResult,
  type ApiTestingNotificationHistory,
  type ApiTestingScanHistory,
} from "../api/client";
import { ResultAsync } from "neverthrow";
import { toError } from "../utils/neverthrowUtils";
import type { paths } from "../api/schema";

export type DeliveryMode =
  | "normal"
  | "guildChannelOnly"
  | "subscribedDMs"
  | "specificUserDM";

export type TestingNotificationRequestBody =
  paths["/api/testing/notifications"]["post"]["requestBody"]["content"]["application/json"];

export type TestingScanRequestBody =
  paths["/api/testing/scans/run"]["post"]["requestBody"]["content"]["application/json"];

function getClientWithHost() {
  const host = getApiHost();
  if (host) {
    const { default: createClient } = { default: apiClient };
    void createClient;
    return apiClient;
  }
  return apiClient;
}

export function useTestingConsole() {
  const capabilities = ref<ApiTestingCapabilities | null>(null);
  const notificationHistory = ref<ApiTestingNotificationHistory["results"]>([]);
  const scanHistory = ref<ApiTestingScanHistory["results"]>([]);
  const lastNotificationResult = ref<ApiTestingNotificationResult | null>(null);
  const lastScanResult = ref<ApiTestingScanResult | null>(null);
  const loading = ref(false);
  const sending = ref(false);
  const scanning = ref(false);
  const errorMessage = ref<string | null>(null);
  const successMessage = ref<string | null>(null);

  async function loadCapabilities() {
    loading.value = true;
    errorMessage.value = null;

    const result = await ResultAsync.fromThrowable(async () => {
      const res = await apiClient.GET("/api/testing/capabilities", {
        headers: getApiHeaders(),
      });
      return res.data ?? null;
    }, (error) => toError(error, "Failed to load testing capabilities"))();

    if (result.isErr()) {
      errorMessage.value = result.error.message;
    } else {
      capabilities.value = result.value;
    }
    loading.value = false;
  }

  async function refreshHistory() {
    const [notifResult, scanResult] = await Promise.all([
      ResultAsync.fromThrowable(async () => {
        const res = await apiClient.GET("/api/testing/notifications", {
          headers: getApiHeaders(),
        });
        return res.data?.results ?? [];
      }, (error) => toError(error, "Failed to load notification history"))(),
      ResultAsync.fromThrowable(async () => {
        const res = await apiClient.GET("/api/testing/runs", {
          headers: getApiHeaders(),
        });
        return res.data?.results ?? [];
      }, (error) => toError(error, "Failed to load scan history"))(),
    ]);

    if (notifResult.isOk()) notificationHistory.value = notifResult.value;
    if (scanResult.isOk()) scanHistory.value = scanResult.value;
  }

  async function sendTestNotification(
    payload: TestingNotificationRequestBody,
  ) {
    sending.value = true;
    errorMessage.value = null;
    successMessage.value = null;

    const result = await ResultAsync.fromThrowable(async () => {
      const res = await apiClient.POST("/api/testing/notifications", {
        body: payload,
        headers: getApiHeaders(),
      });
      return res.data ?? null;
    }, (error) => toError(error, "Failed to send test notification"))();

    if (result.isErr()) {
      errorMessage.value = result.error.message;
    } else {
      lastNotificationResult.value = result.value;
      successMessage.value = "Test notification sent";
      await refreshHistory();
    }

    sending.value = false;
    return result;
  }

  async function runTestScan(
    payload: TestingScanRequestBody,
  ) {
    scanning.value = true;
    errorMessage.value = null;
    successMessage.value = null;

    const result = await ResultAsync.fromThrowable(async () => {
      const res = await apiClient.POST("/api/testing/scans/run", {
        body: payload,
        headers: getApiHeaders(),
      });
      return res.data ?? null;
    }, (error) => toError(error, "Failed to run test scan"))();

    if (result.isErr()) {
      errorMessage.value = result.error.message;
    } else {
      lastScanResult.value = result.value;
      successMessage.value = `Scan completed: ${result.value?.notifications?.length ?? 0} notification(s)`;
      await refreshHistory();
    }

    scanning.value = false;
    return result;
  }

  return {
    capabilities,
    notificationHistory,
    scanHistory,
    lastNotificationResult,
    lastScanResult,
    loading,
    sending,
    scanning,
    errorMessage,
    successMessage,
    loadCapabilities,
    refreshHistory,
    sendTestNotification,
    runTestScan,
    getClientWithHost,
  };
}
