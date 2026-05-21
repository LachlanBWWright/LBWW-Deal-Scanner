<script setup lang="ts">
import { computed } from "vue";
import { ClipboardCopy, PencilLine, Trash2 } from "lucide-vue-next";
import { ResultAsync } from "neverthrow";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { QueryItem, QueryType, SearchResultItem } from "./queryTypes";
import { formatQueryTypeLabel } from "./queryTypes";
import { formatDisplayLabel } from "../utils/displayLabels";

const props = defineProps<{
  filteredQueries: QueryItem[];
  searchResults: SearchResultItem[];
  editingQueryId: string | null;
  editingQueryType: QueryType | null;
}>();

const emit = defineEmits<{
  (e: "edit", item: QueryItem): void;
  (e: "delete", item: QueryItem): void;
}>();

function queryTitle(item: QueryItem) {
  return item.name || item.url || item.displayUrl || item.id;
}

function queryUrl(item: QueryItem) {
  return item.url || item.displayUrl || null;
}

function queryMeta(item: QueryItem) {
  const entries: string[] = [];
  if (item.maxPrice != null) entries.push(`max $${item.maxPrice}`);
  if (item.minPrice != null) entries.push(`min $${item.minPrice}`);
  if (item.minFloat != null) entries.push(`min float ${item.minFloat}`);
  if (item.maxFloat != null) entries.push(`max float ${item.maxFloat}`);
  if (item.dmOnly) entries.push("DM only");
  return entries;
}

function isEditing(item: QueryItem) {
  return props.editingQueryId === item.id && props.editingQueryType === item.type;
}

function queryKey(item: QueryItem) {
  return `${item.type}:${item.id}`;
}

function resultKey(result: SearchResultItem) {
  if (!result.queryType || !result.queryId) return null;
  return `${result.queryType}:${result.queryId}`;
}

function resultTimestamp(result: SearchResultItem) {
  const timestamp = Date.parse(result.foundAt);
  return Number.isNaN(timestamp) ? 0 : timestamp;
}

type LatestResultDisplay = {
  title: string;
  sourceLabel: string;
  foundAtLabel: string;
  price: number | null;
  url: string;
};

const latestResultByQuery = computed(() => {
  const latest = new Map<string, SearchResultItem>();

  for (const result of props.searchResults) {
    const key = resultKey(result);
    if (!key) continue;

    const existing = latest.get(key);
    if (!existing || resultTimestamp(result) > resultTimestamp(existing)) {
      latest.set(key, result);
    }
  }

  return latest;
});

function latestResultFor(item: QueryItem): LatestResultDisplay | null {
  const result = latestResultByQuery.value.get(queryKey(item));
  if (!result) return null;

  return {
    title: result.title,
    sourceLabel: formatDisplayLabel(result.source),
    foundAtLabel: new Date(result.foundAt).toLocaleString(),
    price: result.price,
    url: result.url,
  };
}

function copyQueryUrl(item: QueryItem) {
  const url = queryUrl(item);
  if (!url || !navigator.clipboard) return;

  void ResultAsync.fromPromise(
    navigator.clipboard.writeText(url),
    () => null,
  );
}
</script>
<script lang="ts">
export default {};
</script>

<template>
  <div>
    <div class="mb-3 mt-4 flex items-center justify-between gap-3">
      <h2 class="text-lg font-semibold">Saved queries</h2>
      <p class="text-sm text-muted-foreground">{{ props.filteredQueries.length }} saved queries</p>
    </div>

    <p
      v-if="props.filteredQueries.length === 0"
      class="rounded-md border border-dashed border-border/80 bg-muted/30 px-3 py-4 text-sm text-muted-foreground"
    >
      No queries match the current filters.
    </p>

    <ul class="space-y-3">
      <li
        v-for="item in props.filteredQueries"
        :key="`${item.type}:${item.id}`"
        :class="[
          'rounded-xl border bg-card p-4 shadow-sm transition-colors',
          isEditing(item) ? 'border-sky-300 bg-sky-50/70' : 'border-border/70',
        ]"
      >
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="min-w-0 flex-1 space-y-3">
            <div class="flex flex-wrap items-center gap-2">
              <strong class="max-w-full truncate">{{ queryTitle(item) }}</strong>
              <Badge
                v-if="isEditing(item)"
                variant="outline"
                class="border-sky-300 bg-sky-100 text-sky-800"
              >
                Editing
              </Badge>
            </div>

            <p class="text-xs text-muted-foreground">
              {{ formatQueryTypeLabel(item.type) }} / {{ item.id }}
            </p>

            <div class="space-y-1">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <p class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">
                  Latest item found
                </p>
                <p v-if="latestResultFor(item)" class="text-xs text-muted-foreground">
                  {{ latestResultFor(item)?.foundAtLabel }}
                </p>
              </div>

              <div v-if="latestResultFor(item)" class="space-y-1">
                <p class="text-sm font-semibold text-foreground">
                  {{ latestResultFor(item)?.title ?? "" }}
                </p>
                <p class="text-xs text-muted-foreground">
                  Source: {{ latestResultFor(item)?.sourceLabel ?? "" }}
                  <span v-if="latestResultFor(item)?.price != null">
                    · Price: {{ latestResultFor(item)?.price }}
                  </span>
                </p>
              </div>

              <p v-else class="text-sm text-muted-foreground">
                No recent result has been recorded for this query yet.
              </p>
            </div>

            <div v-if="queryMeta(item).length" class="flex flex-wrap gap-2">
              <Badge v-for="entry in queryMeta(item)" :key="entry" variant="secondary">
                {{ entry }}
              </Badge>
            </div>
          </div>

          <div class="flex shrink-0 flex-col items-end gap-2">
            <Button
              variant="outline"
              size="icon"
              type="button"
              class="h-9 w-9"
              title="Copy query URL"
              aria-label="Copy query URL"
              :disabled="!queryUrl(item)"
              @click="copyQueryUrl(item)"
            >
              <ClipboardCopy class="h-4 w-4" />
            </Button>
            <Button
              variant="secondary"
              size="icon"
              type="button"
              class="h-9 w-9"
              title="Edit query"
              aria-label="Edit query"
              @click="emit('edit', item)"
            >
              <PencilLine class="h-4 w-4" />
            </Button>
            <Button
              variant="destructive"
              size="icon"
              type="button"
              class="h-9 w-9"
              title="Delete query"
              aria-label="Delete query"
              @click="emit('delete', item)"
            >
              <Trash2 class="h-4 w-4" />
            </Button>
          </div>
        </div>
      </li>
    </ul>
  </div>
</template>
