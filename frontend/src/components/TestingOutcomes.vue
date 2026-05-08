<script setup lang="ts">
import { computed, ref } from "vue";
import { ResultAsync } from "neverthrow";
import { useTestingConsole } from "../composables/useTestingConsole";
import { apiClient, getApiHeaders } from "../api/client";
import { toError } from "../utils/neverthrowUtils";
import type { paths } from "../api/schema";

type SearchResultItem =
  paths["/api/search-results"]["get"]["responses"][200]["content"]["application/json"]["results"][number];

const {
  notificationHistory,
  scanHistory,
  refreshHistory,
} = useTestingConsole();

const recentResults = ref<SearchResultItem[]>([]);
const loadingResults = ref(false);
const resultsError = ref<string | null>(null);

type OutcomeFilter = "all" | "sent" | "failed";
const outcomeFilter = ref<OutcomeFilter>("all");
const sourceFilter = ref("");

const filteredNotifications = computed(() => {
  return notificationHistory.value.filter((r) => {
    if (
      outcomeFilter.value !== "all" &&
      !r.outcomes.some((o) => o.status === outcomeFilter.value)
    ) {
      return false;
    }
    return true;
  });
});

const filteredScans = computed(() => {
  return scanHistory.value.filter((r) => {
    if (
      outcomeFilter.value === "failed" &&
      r.errors.length === 0
    ) {
      return false;
    }
    if (
      outcomeFilter.value === "sent" &&
      !r.notificationsPublished
    ) {
      return false;
    }
    return true;
  });
});

const filteredRecentResults = computed(() => {
  const sf = sourceFilter.value.trim().toLowerCase();
  if (!sf) return recentResults.value;
  return recentResults.value.filter((r) => r.source.toLowerCase().includes(sf));
});

async function loadRecentResults() {
  loadingResults.value = true;
  resultsError.value = null;

  const result = await ResultAsync.fromThrowable(async () => {
    const res = await apiClient.GET("/api/search-results", {
      headers: getApiHeaders(),
    });
    return res.data?.results ?? [];
  }, (error) => toError(error, "Failed to load recent results"))();

  if (result.isOk()) {
    recentResults.value = result.value;
  } else {
    resultsError.value = result.error.message;
  }
  loadingResults.value = false;
}

async function refresh() {
  await Promise.all([refreshHistory(), loadRecentResults()]);
}

function formatTime(iso: string) {
  return new Date(iso).toLocaleTimeString();
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString();
}
</script>
<script lang="ts">
export default {};
</script>

<template>
  <div class="outcomes-panel">
    <div class="outcomes-header">
      <h3>Outcomes</h3>
      <div class="filter-row">
        <select v-model="outcomeFilter" class="filter-select">
          <option value="all">All outcomes</option>
          <option value="sent">Sent / published</option>
          <option value="failed">Failed / errors</option>
        </select>
        <input
          v-model="sourceFilter"
          class="filter-input"
          type="text"
          placeholder="Filter by source…"
        />
        <button class="refresh-btn secondary" @click="refresh">Refresh</button>
      </div>
    </div>

    <!-- Notification history -->
    <section class="outcome-section">
      <h4>Artificial notification history ({{ filteredNotifications.length }})</h4>
      <p v-if="!filteredNotifications.length" class="empty-msg">No entries</p>
      <div
        v-for="entry in filteredNotifications"
        :key="entry.id"
        class="outcome-card"
      >
        <div class="oc-row">
          <span class="oc-time">{{ formatTime(entry.startedAt) }}</span>
          <span class="oc-dur">{{ entry.durationMs }}ms</span>
          <span
            v-for="o in entry.outcomes"
            :key="o.provider"
            :class="{
              'outcome-sent': o.status === 'sent',
              'outcome-failed': o.status === 'failed',
              'outcome-skip': o.status === 'skipped' || o.status === 'disabled',
            }"
          >{{ o.status }}</span>
        </div>
        <p v-if="entry.error" class="oc-error">{{ entry.error }}</p>
      </div>
    </section>

    <!-- Scan history -->
    <section class="outcome-section">
      <h4>Temporary scan history ({{ filteredScans.length }})</h4>
      <p v-if="!filteredScans.length" class="empty-msg">No entries</p>
      <div
        v-for="entry in filteredScans"
        :key="entry.id"
        class="outcome-card"
      >
        <div class="oc-row">
          <span class="oc-time">{{ formatTime(entry.startedAt) }}</span>
          <span class="oc-dur">{{ entry.durationMs }}ms</span>
          <span class="oc-notifs">{{ entry.notifications.length }} notification(s)</span>
          <span v-if="entry.notificationsPublished" class="outcome-sent">published</span>
          <span v-if="entry.errors.length" class="outcome-failed">{{ entry.errors.length }} error(s)</span>
        </div>
        <div v-if="entry.notifications.length" class="oc-notif-list">
          <a
            v-for="n in entry.notifications"
            :key="n.url"
            :href="n.url"
            target="_blank"
            rel="noopener noreferrer"
            class="oc-notif-link"
          >{{ n.title }}</a>
        </div>
        <div v-if="entry.errors.length" class="oc-errors">
          <p v-for="(err, idx) in entry.errors" :key="idx" class="oc-error">{{ err }}</p>
        </div>
      </div>
    </section>

    <!-- Recent scanner results -->
    <section class="outcome-section">
      <h4>Recent scanner results</h4>
      <p v-if="resultsError" class="oc-error">{{ resultsError }}</p>
      <button v-if="!recentResults.length && !loadingResults" class="secondary small-btn" @click="loadRecentResults">
        Load results
      </button>
      <p v-if="loadingResults" class="empty-msg">Loading…</p>
      <p v-if="!loadingResults && !recentResults.length && !resultsError" class="empty-msg">
        Click "Load results" or "Refresh" to fetch
      </p>
      <div class="results-table" v-if="filteredRecentResults.length">
        <div class="results-header-row">
          <span>Source</span>
          <span>Title</span>
          <span>Price</span>
          <span>Found</span>
        </div>
        <div
          v-for="r in filteredRecentResults"
          :key="r.url + r.foundAt"
          class="results-data-row"
        >
          <span class="badge-source">{{ r.source }}</span>
          <a :href="r.url" target="_blank" rel="noopener noreferrer" class="result-title-link">{{ r.title }}</a>
          <span>{{ r.price !== null ? `$${r.price}` : "—" }}</span>
          <span class="result-time">{{ formatDate(r.foundAt) }}</span>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.outcomes-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.outcomes-header {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.outcomes-header h3 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
}

.filter-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.filter-select,
.filter-input {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(154, 173, 201, 0.2);
  border-radius: 6px;
  padding: 5px 8px;
  color: inherit;
  font: inherit;
  font-size: 0.8rem;
}

.refresh-btn {
  padding: 5px 12px;
  font-size: 0.8rem;
  border-radius: 6px;
}

.small-btn {
  padding: 4px 10px;
  font-size: 0.78rem;
  border-radius: 6px;
}

.outcome-section h4 {
  margin: 0 0 8px;
  font-size: 0.82rem;
  font-weight: 600;
  color: rgba(219, 227, 240, 0.6);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.empty-msg {
  font-size: 0.8rem;
  color: rgba(219, 227, 240, 0.4);
  font-style: italic;
  margin: 0;
}

.outcome-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(154, 173, 201, 0.12);
  border-radius: 6px;
  padding: 6px 10px;
  margin-bottom: 4px;
  font-size: 0.8rem;
}

.oc-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.oc-time {
  color: rgba(219, 227, 240, 0.45);
  font-size: 0.75rem;
}

.oc-dur {
  color: rgba(219, 227, 240, 0.45);
  font-size: 0.75rem;
}

.oc-notifs {
  color: rgba(219, 227, 240, 0.7);
}

.outcome-sent { color: #82ffb4; }
.outcome-failed { color: #ff8080; }
.outcome-skip { color: #ffa064; }

.oc-error {
  color: #ff8080;
  font-size: 0.78rem;
  margin: 4px 0 0;
}

.oc-notif-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-top: 4px;
}

.oc-notif-link {
  color: #7dfdd4;
  font-size: 0.78rem;
  text-decoration: none;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.oc-notif-link:hover {
  text-decoration: underline;
}

.oc-errors {
  margin-top: 4px;
}

.results-table {
  display: grid;
  grid-template-columns: 100px 1fr 70px 120px;
  gap: 2px 8px;
  font-size: 0.78rem;
}

.results-header-row {
  display: contents;
  font-size: 0.72rem;
  font-weight: 600;
  color: rgba(219, 227, 240, 0.4);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.results-header-row span {
  padding: 0 0 4px;
  border-bottom: 1px solid rgba(154, 173, 201, 0.15);
}

.results-data-row {
  display: contents;
}

.results-data-row > span,
.results-data-row > a {
  padding: 4px 0;
  border-bottom: 1px solid rgba(154, 173, 201, 0.06);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.badge-source {
  font-size: 0.68rem;
  padding: 1px 5px;
  border-radius: 3px;
  background: rgba(62, 167, 255, 0.15);
  color: #3ea7ff;
  border: 1px solid rgba(62, 167, 255, 0.2);
  display: inline-block;
}

.result-title-link {
  color: #c8d6ef;
  text-decoration: none;
}

.result-title-link:hover {
  color: #7dfdd4;
  text-decoration: underline;
}

.result-time {
  color: rgba(219, 227, 240, 0.4);
  font-size: 0.72rem;
}
</style>
