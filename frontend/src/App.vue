<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref } from "vue";
import { ResultAsync } from "neverthrow";

import {
  apiClient,
  getApiHeaders,
  getApiHost,
  setApiHost,
  getApiSecret,
  setApiSecret,
  type ApiStatus,
} from "./api/client";
import QueryManager from "./components/QueryManager.vue";
import { toError } from "./utils/neverthrowUtils";
import { useTestingConsole } from "./composables/useTestingConsole";

const NotificationTester = defineAsyncComponent(
  () => import("./components/NotificationTester.vue"),
);
const ManualScanTester = defineAsyncComponent(
  () => import("./components/ManualScanTester.vue"),
);
const TestingOutcomes = defineAsyncComponent(
  () => import("./components/TestingOutcomes.vue"),
);

type Tab = "queries" | "scans" | "notifications" | "outcomes";
const activeTab = ref<Tab>("queries");
const tabs: Tab[] = ["queries", "scans", "notifications", "outcomes"];

const runtime = ref<ApiStatus | null>(null);
const loading = ref(false);
const errorMessage = ref<string | null>(null);
const apiHost = ref(getApiHost());
const apiSecret = ref(getApiSecret());
const showSettings = ref(false);

const { capabilities, loadCapabilities } = useTestingConsole();

const botStatus = computed(() => {
  const bot = runtime.value?.bot;
  if (!bot) return { label: "—", dot: "dot-amber" };
  if (bot.connected) return { label: "Connected", dot: "dot-lime" };
  return { label: bot.status, dot: "dot-amber" };
});

const scanStatus = computed(() => {
  const scanner = runtime.value?.scanner;
  if (!scanner) return { label: "—", dot: "dot-amber" };
  if (scanner.loopRunning) return { label: "Running", dot: "dot-lime" };
  return { label: "Idle", dot: "dot-amber" };
});

const testingBadge = computed(() =>
  capabilities.value?.testingEnabled ? "Testing ON" : "Testing OFF",
);

function saveSettings() {
  setApiHost(apiHost.value);
  setApiSecret(apiSecret.value);
  showSettings.value = false;
  void Promise.all([refresh(), loadCapabilities()]);
}

async function refresh() {
  loading.value = true;
  errorMessage.value = null;

  const result = await ResultAsync.fromThrowable(async () => {
    const res = await apiClient.GET("/api/status", { headers: getApiHeaders() });
    return res.data ?? null;
  }, (error) => toError(error, "Failed to load status"))();

  if (result.isErr()) {
    errorMessage.value = result.error.message;
  } else {
    runtime.value = result.value;
  }
  loading.value = false;
}

onMounted(async () => {
  await Promise.all([refresh(), loadCapabilities()]);
});
</script>

<template>
  <div class="console-shell">
    <!-- Top bar -->
    <header class="top-bar">
      <span class="app-name">DealScanner</span>

      <div class="status-chips">
        <span class="chip">
          <span class="dot" :class="botStatus.dot" />
          Bot: {{ botStatus.label }}
        </span>
        <span class="chip">
          <span class="dot" :class="scanStatus.dot" />
          Scan: {{ scanStatus.label }}
        </span>
        <span
          class="chip"
          :class="capabilities?.testingEnabled ? 'chip-green' : 'chip-muted'"
        >
          {{ testingBadge }}
        </span>
      </div>

      <div class="top-actions">
        <button
          class="icon-btn"
          :disabled="loading"
          title="Refresh status"
          @click="refresh"
        >
          {{ loading ? "…" : "↻" }}
        </button>
        <button class="icon-btn" title="API settings" @click="showSettings = !showSettings">
          ⚙
        </button>
      </div>
    </header>

    <!-- Settings drawer -->
    <div v-if="showSettings" class="settings-drawer">
      <label class="setting-field">
        <span>API host</span>
        <input v-model="apiHost" type="text" placeholder="Leave empty for current host" />
      </label>
      <label class="setting-field">
        <span>API secret</span>
        <input v-model="apiSecret" type="password" placeholder="••••••••" />
      </label>
      <div class="setting-actions">
        <button class="primary" @click="saveSettings">Save</button>
        <button class="secondary" @click="showSettings = false">Cancel</button>
      </div>
      <p v-if="errorMessage" class="setting-error">{{ errorMessage }}</p>
    </div>

    <!-- Tab bar -->
    <nav class="tab-bar">
      <button
        v-for="tab in tabs"
        :key="tab"
        class="tab-btn"
        :class="{ active: activeTab === tab }"
        @click="activeTab = tab"
      >
        {{
          tab === "queries" ? "Queries"
          : tab === "scans" ? "Manual scans"
          : tab === "notifications" ? "Notifications"
          : "Outcomes"
        }}
      </button>
    </nav>

    <!-- Tab content -->
    <main class="console-content">
      <div v-if="activeTab === 'queries'">
        <QueryManager />
      </div>

      <div v-if="activeTab === 'scans'" class="tab-pane">
        <ManualScanTester />
      </div>

      <div v-if="activeTab === 'notifications'" class="tab-pane">
        <NotificationTester />
      </div>

      <div v-if="activeTab === 'outcomes'" class="tab-pane">
        <TestingOutcomes />
      </div>
    </main>
  </div>
</template>
