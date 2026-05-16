<script setup lang="ts">
import type { SearchResultItem } from "./queryTypes";

const props = defineProps<{
  searchResults: SearchResultItem[];
  loading: boolean;
}>();

const emit = defineEmits<{
  (e: "refresh"): void;
}>();

function formatFoundAt(foundAt: string) {
  return new Date(foundAt).toLocaleString();
}
</script>
<script lang="ts">
export default {};
</script>

<template>
  <div class="results-section">
    <div class="panel-head results-head">
      <h2>Recent search results</h2>
      <span>{{ props.searchResults.length }} results</span>
    </div>

    <div class="query-toolbar">
      <button class="secondary" type="button" @click="emit('refresh')" :disabled="props.loading">
        Refresh results
      </button>
    </div>

    <p v-if="!props.loading && props.searchResults.length === 0" class="empty-message">
      No results have been recorded yet. Run a scan to populate this list.
    </p>

    <ul class="query-list">
      <li
        v-for="(result, index) in props.searchResults"
        :key="`${result.source}:${result.queryId ?? result.url}:${result.foundAt}:${index}`"
        class="query-card"
      >
        <div class="query-header">
          <strong>{{ result.source }}</strong>
          <span>{{ formatFoundAt(result.foundAt) }}</span>
        </div>

        <p class="result-title">{{ result.title }}</p>

        <div class="result-meta">
          <span v-if="result.price !== null">Price: {{ result.price }}</span>
          <span v-if="result.queryType">Query type: {{ result.queryType }}</span>
          <span v-if="result.queryId">Query id: {{ result.queryId }}</span>
        </div>

        <a :href="result.url" target="_blank" rel="noopener noreferrer" class="result-link">
          {{ result.url }}
        </a>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.results-section {
  margin-top: 18px;
}

.results-head {
  margin-top: 0;
}

.result-title {
  margin: 0 0 8px;
  color: rgba(238, 243, 251, 0.92);
  font-weight: 600;
}

.result-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 8px;
  color: rgba(215, 222, 234, 0.68);
  font-size: 0.8rem;
}

.result-link {
  color: #79b8ff;
  word-break: break-all;
  font-size: 0.8rem;
}

.empty-message {
  margin: 14px 0 0;
  color: rgba(215, 222, 234, 0.66);
}
</style>
