<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import {
  apiClient,
  type ApiRunScanResponse,
  type ApiStatus,
} from "./api/client";

const runtime = ref<ApiStatus | null>(null);
const commands = ref<string[]>([]);
const lastRun = ref<ApiRunScanResponse | null>(null);
const loading = ref(false);
const runningScan = ref(false);
const errorMessage = ref<string | null>(null);

const stats = computed(() => {
  const scanner = runtime.value?.scanner;
  return [
    {
      label: "API port",
      value: runtime.value ? String(runtime.value.api.port) : "—",
      tone: "cyan",
    },
    {
      label: "Scan loop",
      value: scanner?.loopRunning ? "Running" : "Idle",
      tone: scanner?.loopRunning ? "lime" : "amber",
    },
    {
      label: "Discord bot",
      value: runtime.value?.bot.connected
        ? "Connected"
        : runtime.value?.bot.status ?? "Waiting",
      tone: runtime.value?.bot.connected ? "lime" : "amber",
    },
  ];
});

const shortcuts = computed(() => [
  {
    title: "Run manual scan",
    detail: "Triggers a one-off pass through the scanner set.",
    action: "scan" as const,
  },
  {
    title: "Refresh runtime",
    detail: "Pull the latest API, bot, and scanner status.",
    action: "refresh" as const,
  },
  {
    title: "Open Swagger UI",
    detail: "Inspect the generated Fastify contract in the browser.",
    action: "docs" as const,
  },
]);

const activity = computed(() => {
  const items = [
    runtime.value?.bot.lastReadyAt
      ? `Bot ready since ${new Date(runtime.value.bot.lastReadyAt).toLocaleString()}`
      : "Bot status has not been reported yet",
    lastRun.value
      ? `Last manual scan finished in ${Math.round(lastRun.value.durationMs / 1000)}s`
      : "No manual scan has been triggered from the UI yet",
    `Registered commands: ${commands.value.length}`,
  ];

  if (runtime.value?.scanner.lastRun?.error) {
    items.unshift(`Last scan warning: ${runtime.value.scanner.lastRun.error}`);
  }

  return items;
});

async function refreshRuntime() {
  loading.value = true;
  errorMessage.value = null;

  try {
    const [statusResponse, commandsResponse] = await Promise.all([
      apiClient.GET("/api/status"),
      apiClient.GET("/api/commands"),
    ]);

    runtime.value = statusResponse.data ?? null;
    commands.value = commandsResponse.data?.commands ?? [];
  } catch (error) {
    errorMessage.value =
      error instanceof Error ? error.message : "Failed to load runtime state.";
  } finally {
    loading.value = false;
  }
}

async function triggerManualScan() {
  runningScan.value = true;
  errorMessage.value = null;

  try {
    const response = await apiClient.POST("/api/scans/run", {
      body: { reason: "manual-ui" },
    });
    lastRun.value = response.data ?? null;
    await refreshRuntime();
  } catch (error) {
    errorMessage.value =
      error instanceof Error ? error.message : "Manual scan failed.";
  } finally {
    runningScan.value = false;
  }
}

function openDocs() {
  window.open("/docs", "_blank", "noopener,noreferrer");
}

function runShortcut(action: "scan" | "refresh" | "docs") {
  if (action === "scan") {
    void triggerManualScan();
    return;
  }

  if (action === "docs") {
    openDocs();
    return;
  }

  void refreshRuntime();
}

onMounted(() => {
  void refreshRuntime();
});
</script>

<template>
  <main class="shell">
    <section class="hero">
      <div class="hero-copy">
        <p class="eyebrow">DealScanner</p>
        <h1>Local control room for scans, alerts, and bot operations.</h1>
        <p class="lede">
          Use this dashboard to orchestrate scan runs, inspect queued findings,
          and call the generated Fastify API contract directly from Vue.
        </p>
        <div class="hero-actions">
          <button class="primary" :disabled="runningScan" @click="triggerManualScan">
            {{ runningScan ? "Running scan..." : "Start scan cycle" }}
          </button>
          <button class="secondary" :disabled="loading" @click="refreshRuntime">
            {{ loading ? "Refreshing..." : "Refresh status" }}
          </button>
        </div>
        <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
      </div>

      <aside class="status-panel">
        <div class="status-card" v-for="stat in stats" :key="stat.label">
          <span class="status-dot" :class="stat.tone" />
          <div>
            <strong>{{ stat.value }}</strong>
            <p>{{ stat.label }}</p>
          </div>
        </div>
      </aside>
    </section>

    <section class="grid">
      <article class="panel">
        <div class="panel-head">
          <h2>Quick actions</h2>
          <span>{{ commands.length }} commands</span>
        </div>
        <div class="action-list">
          <button
            v-for="shortcut in shortcuts"
            :key="shortcut.title"
            class="action-card"
            :disabled="shortcut.action === 'scan' && runningScan"
            @click="runShortcut(shortcut.action)"
          >
            <strong>{{ shortcut.title }}</strong>
            <p>{{ shortcut.detail }}</p>
          </button>
          <div class="command-list">
            <span class="command-chip" v-for="command in commands" :key="command">
              {{ command }}
            </span>
          </div>
        </div>
      </article>

      <article class="panel">
        <div class="panel-head">
          <h2>Live activity</h2>
          <span>{{ lastRun ? "Synced" : "Idle" }}</span>
        </div>
        <ul class="activity-list">
          <li v-for="item in activity" :key="item">
            <span class="activity-bullet" />
            <span>{{ item }}</span>
          </li>
        </ul>
      </article>
    </section>
  </main>
</template>
