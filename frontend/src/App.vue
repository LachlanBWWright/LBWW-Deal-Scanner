<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref } from "vue";
import { ResultAsync } from "neverthrow";

import {
  getApiClient,
  getApiHeaders,
  getApiHost,
  setApiHost,
  getApiSecret,
  setApiSecret,
  type ApiStatus,
} from "./api/client";
import QueryManager from "./components/QueryManager.vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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

function tabLabel(tab: Tab) {
  if (tab === "queries") return "Queries";
  if (tab === "scans") return "Manual scans";
  if (tab === "notifications") return "Notifications";
  return "Outcomes";
}

const runtime = ref<ApiStatus | null>(null);
const loading = ref(false);
const errorMessage = ref<string | null>(null);
const apiHost = ref(getApiHost());
const apiSecret = ref(getApiSecret());
const showSettings = ref(false);

const {
  capabilities,
  loading: testingLoading,
  loadCapabilities,
} = useTestingConsole();

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

const testingBadge = computed(() => {
  if (testingLoading.value) return "Testing loading...";
  return capabilities.value?.testingEnabled ? "Testing ON" : "Testing OFF";
});

const testingBadgeVariant = computed(() => {
  if (testingLoading.value) return "secondary";
  return capabilities.value?.testingEnabled ? "success" : "warning";
});

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
    const res = await getApiClient().GET("/api/status", {
      headers: getApiHeaders(),
    });
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
  <div class="flex min-h-dvh w-full flex-col">
    <header class="sticky top-0 z-30 border-b border-border/80 bg-slate-950/90 backdrop-blur">
      <div class="mx-auto flex w-full max-w-7xl flex-col gap-1 px-3 py-1 sm:px-4">
        <div class="flex min-h-0 flex-wrap items-center gap-2">
          <nav class="flex flex-1 flex-wrap gap-2 overflow-x-auto">
            <Button
              v-for="tab in tabs"
              :key="tab"
              :variant="activeTab === tab ? 'default' : 'secondary'"
              size="sm"
              class="h-8 px-3"
              @click="activeTab = tab"
            >
              {{ tabLabel(tab) }}
            </Button>
          </nav>

          <div class="ml-auto flex flex-wrap items-center gap-2">
            <Badge variant="outline">Bot: {{ botStatus.label }}</Badge>
            <Badge variant="outline">Scan: {{ scanStatus.label }}</Badge>
            <Badge :variant="testingBadgeVariant">{{ testingBadge }}</Badge>
            <Button size="icon" variant="outline" :disabled="loading" title="Refresh status" @click="refresh">
              {{ loading ? "…" : "↻" }}
            </Button>
            <Button size="icon" variant="outline" title="API settings" @click="showSettings = !showSettings">
              ⚙
            </Button>
          </div>
        </div>

        <div v-if="showSettings" class="grid gap-3 rounded-lg border border-border/70 bg-card/70 p-3 md:grid-cols-3">
          <div class="space-y-2 md:col-span-1">
            <Label>API host</Label>
            <Input v-model="apiHost" type="text" placeholder="Leave empty for current host" />
          </div>
          <div class="space-y-2 md:col-span-1">
            <Label>API secret</Label>
            <Input v-model="apiSecret" type="password" placeholder="••••••••" />
          </div>
          <div class="flex items-end gap-2 md:col-span-1">
            <Button @click="saveSettings">Save</Button>
            <Button variant="secondary" @click="showSettings = false">Cancel</Button>
          </div>
          <p v-if="errorMessage" class="md:col-span-3 rounded-md border border-red-500/40 bg-red-500/10 px-3 py-2 text-sm font-medium text-red-200">{{ errorMessage }}</p>
        </div>
      </div>
    </header>

    <main class="flex min-h-0 flex-1">
      <div class="mx-auto flex min-h-0 w-full max-w-7xl flex-1 px-3 py-2 sm:px-4">
        <div class="flex min-h-0 w-full flex-1">
          <QueryManager v-if="activeTab === 'queries'" :apiHost="apiHost" :apiSecret="apiSecret" />
          <div v-if="activeTab === 'scans'" class="mx-auto flex min-h-0 w-full max-w-5xl flex-1">
            <ManualScanTester />
          </div>
          <div v-if="activeTab === 'notifications'" class="mx-auto flex min-h-0 w-full max-w-5xl flex-1">
            <NotificationTester />
          </div>
          <div v-if="activeTab === 'outcomes'" class="mx-auto flex min-h-0 w-full max-w-5xl flex-1">
            <TestingOutcomes />
          </div>
        </div>
      </div>
    </main>
  </div>
</template>
