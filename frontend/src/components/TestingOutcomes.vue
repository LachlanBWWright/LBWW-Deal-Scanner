<script setup lang="ts">
import { computed, ref } from "vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { ResultAsync } from "neverthrow";
import { useTestingConsole } from "../composables/useTestingConsole";
import { getApiClient, getApiHeaders } from "../api/client";
import { toError } from "../utils/neverthrowUtils";
import type { paths } from "../api/schema";
import { formatDisplayLabel } from "../utils/displayLabels";

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
    if (outcomeFilter.value !== "all" && !r.outcomes.some((o) => o.status === outcomeFilter.value)) {
      return false;
    }
    return true;
  });
});

const filteredScans = computed(() => {
  return scanHistory.value.filter((r) => {
    if (outcomeFilter.value === "failed" && r.errors.length === 0) {
      return false;
    }
    if (outcomeFilter.value === "sent" && !r.notificationsPublished) {
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

function updateOutcomeFilter(value: string) {
  if (value === "all" || value === "sent" || value === "failed") {
    outcomeFilter.value = value;
  }
}

async function loadRecentResults() {
  loadingResults.value = true;
  resultsError.value = null;

  const result = await ResultAsync.fromThrowable(async () => {
    const res = await getApiClient().GET("/api/search-results", {
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
  <Card class="flex min-h-0 flex-1 flex-col">
    <CardContent class="flex min-h-0 flex-1 flex-col space-y-6 pt-6">
      <div class="flex flex-wrap items-center gap-2">
        <h2 class="text-lg font-semibold tracking-tight">Outcomes</h2>
        <Select :model-value="outcomeFilter" class="w-[180px]" @update:model-value="updateOutcomeFilter">
          <option value="all">All outcomes</option>
          <option value="sent">Sent / published</option>
          <option value="failed">Failed / errors</option>
        </Select>
        <Input v-model="sourceFilter" class="w-[220px]" type="text" placeholder="Filter by source..." />
        <Button variant="secondary" @click="refresh">Refresh</Button>
      </div>

      <section class="space-y-3">
        <h4 class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Artificial notification history ({{ filteredNotifications.length }})</h4>
        <p v-if="!filteredNotifications.length" class="text-sm italic text-muted-foreground">No entries</p>
        <div v-for="entry in filteredNotifications" :key="entry.id" class="space-y-2 rounded-lg border border-border/70 bg-muted/20 p-3 text-sm">
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-xs text-muted-foreground">{{ formatTime(entry.startedAt) }}</span>
            <span class="text-xs text-muted-foreground">{{ entry.durationMs }}ms</span>
            <Badge
              v-for="o in entry.outcomes"
              :key="o.provider"
              :variant="o.status === 'sent' ? 'success' : o.status === 'failed' ? 'destructive' : 'warning'"
            >
              {{ o.status }}
            </Badge>
          </div>
          <p v-if="entry.error" class="text-sm text-red-200">{{ entry.error }}</p>
        </div>
      </section>

      <section class="space-y-3">
        <h4 class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Temporary scan history ({{ filteredScans.length }})</h4>
        <p v-if="!filteredScans.length" class="text-sm italic text-muted-foreground">No entries</p>
        <div v-for="entry in filteredScans" :key="entry.id" class="space-y-2 rounded-lg border border-border/70 bg-muted/20 p-3 text-sm">
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-xs text-muted-foreground">{{ formatTime(entry.startedAt) }}</span>
            <span class="text-xs text-muted-foreground">{{ entry.durationMs }}ms</span>
            <Badge variant="secondary">{{ entry.notifications.length }} notification(s)</Badge>
            <Badge v-if="entry.notificationsPublished" variant="success">published</Badge>
            <Badge v-if="entry.errors.length" variant="destructive">{{ entry.errors.length }} error(s)</Badge>
          </div>
          <div v-if="entry.notifications.length" class="space-y-1">
            <a
              v-for="n in entry.notifications"
              :key="n.url"
              :href="n.url"
              target="_blank"
              rel="noopener noreferrer"
              class="block truncate text-sm font-medium text-blue-300 hover:underline"
            >{{ n.title }}</a>
          </div>
          <div v-if="entry.errors.length" class="space-y-1">
            <p v-for="(err, idx) in entry.errors" :key="idx" class="text-sm text-red-200">{{ err }}</p>
          </div>
        </div>
      </section>

      <section class="space-y-3">
        <h4 class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Recent scanner results</h4>
        <p v-if="resultsError" class="text-sm text-red-200">{{ resultsError }}</p>
        <Button v-if="!recentResults.length && !loadingResults" variant="outline" size="sm" @click="loadRecentResults">
          Load results
        </Button>
        <p v-if="loadingResults" class="text-sm italic text-muted-foreground">Loading...</p>
        <p v-if="!loadingResults && !recentResults.length && !resultsError" class="text-sm italic text-muted-foreground">
          Click "Load results" or "Refresh" to fetch
        </p>
        <div v-if="filteredRecentResults.length" class="overflow-x-auto rounded-lg border border-border/70">
          <table class="min-w-full text-sm">
            <thead class="bg-muted/30 text-left text-xs uppercase tracking-wide text-muted-foreground">
              <tr>
                <th class="px-3 py-2">Source</th>
                <th class="px-3 py-2">Title</th>
                <th class="px-3 py-2">Price</th>
                <th class="px-3 py-2">Found</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in filteredRecentResults" :key="r.url + r.foundAt" class="border-t border-border/70">
                <td class="px-3 py-2"><Badge variant="secondary">{{ formatDisplayLabel(r.source) }}</Badge></td>
                <td class="px-3 py-2"><a :href="r.url" target="_blank" rel="noopener noreferrer" class="text-blue-300 hover:underline">{{ r.title }}</a></td>
                <td class="px-3 py-2">{{ r.price !== null ? `$${r.price}` : "-" }}</td>
                <td class="px-3 py-2 text-xs text-muted-foreground">{{ formatDate(r.foundAt) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </CardContent>
  </Card>
</template>
