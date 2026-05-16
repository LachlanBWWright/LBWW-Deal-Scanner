import { ref } from "vue";
import {
  getApiClient,
  getApiHeaders,
  type ApiTestingCapabilities,
  type ApiTestingNotificationResult,
  type ApiTestingScanResult,
  type ApiTestingNotificationHistory,
  type ApiTestingScanHistory,
} from "../api/client";
import { err, ok, ResultAsync, type Result } from "neverthrow";
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

function apiResponseResult<T>(
  response: { data?: T; error?: unknown },
  context: string,
): Result<T | null, Error> {
  if (response.error !== undefined) {
    return err(toError(response.error, context));
  }
  return ok(response.data ?? null);
}

export function useTestingConsole() {
  async function loadCapabilities() {
    loading.value = true;
    errorMessage.value = null;

    const result = await ResultAsync.fromThrowable(
      () =>
        getApiClient().GET("/api/testing/capabilities", {
          headers: getApiHeaders(),
        }),
      (error) => toError(error, "Failed to load testing capabilities"),
    )()
      .andThen((response) =>
        apiResponseResult(response, "Failed to load testing capabilities"),
      );

    if (result.isErr()) {
      errorMessage.value = result.error.message;
    } else {
      capabilities.value = result.value;
    }
    loading.value = false;
  }

  async function refreshHistory() {
    const [notifResult, scanResult] = await Promise.all([
      ResultAsync.fromThrowable(
        () =>
          getApiClient().GET("/api/testing/notifications", {
            headers: getApiHeaders(),
          }),
        (error) => toError(error, "Failed to load notification history"),
      )()
        .andThen((response) =>
          apiResponseResult(response, "Failed to load notification history"),
        )
        .map((data) => data?.results ?? []),
      ResultAsync.fromThrowable(
        () =>
          getApiClient().GET("/api/testing/runs", {
            headers: getApiHeaders(),
          }),
        (error) => toError(error, "Failed to load scan history"),
      )()
        .andThen((response) =>
          apiResponseResult(response, "Failed to load scan history"),
        )
        .map((data) => data?.results ?? []),
    ]);

    if (notifResult.isOk()) notificationHistory.value = notifResult.value;
    if (scanResult.isOk()) scanHistory.value = scanResult.value;
  }

  async function sendTestNotification(payload: TestingNotificationRequestBody) {
    sending.value = true;
    errorMessage.value = null;
    successMessage.value = null;

    const result = await ResultAsync.fromThrowable(
      () =>
        getApiClient().POST("/api/testing/notifications", {
          body: payload,
          headers: getApiHeaders(),
        }),
      (error) => toError(error, "Failed to send test notification"),
    )()
      .andThen((response) =>
        apiResponseResult(response, "Failed to send test notification"),
      );

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

  async function runTestScan(payload: TestingScanRequestBody) {
    scanning.value = true;
    errorMessage.value = null;
    successMessage.value = null;

    const result = await ResultAsync.fromThrowable(
      () =>
        getApiClient().POST("/api/testing/scans/run", {
          body: payload,
          headers: getApiHeaders(),
        }),
      (error) => toError(error, "Failed to run test scan"),
    )()
      .andThen((response) =>
        apiResponseResult(response, "Failed to run test scan"),
      );

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
  };
}
