<script setup lang="ts">
import type { QueryItem, QueryType } from "./queryTypes";

const props = defineProps<{
  filteredQueries: QueryItem[];
  supportedTypes: QueryType[];
  filterType: QueryType | "all";
  querySearch: string;
  loading: boolean;
}>();

const emit = defineEmits<{
  (e: "update:filterType", value: QueryType | "all"): void;
  (e: "update:querySearch", value: string): void;
  (e: "refresh"): void;
  (e: "edit", item: QueryItem): void;
  (e: "delete", item: QueryItem): void;
}>();

function handleFilterChange(event: Event) {
  const value = (event.target as HTMLSelectElement).value as QueryType | "all";
  emit("update:filterType", value);
}

function handleSearchChange(event: Event) {
  emit("update:querySearch", (event.target as HTMLInputElement).value);
}

function formatQuery(item: QueryItem) {
  return JSON.stringify(item, null, 2);
}

function queryTitle(item: QueryItem) {
  return item.name || item.url || item.displayUrl || item.id;
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
</script>
<script lang="ts">
export default {};
</script>

<template>
  <div>
    <div class="panel-head list-head">
      <h2>Saved queries</h2>
      <span>{{ props.filteredQueries.length }} saved queries</span>
    </div>

    <div class="query-toolbar">
      <label>
        Filter type
        <select :value="props.filterType" @change="handleFilterChange">
          <option value="all">All</option>
          <option
            v-for="type in props.supportedTypes"
            :key="type"
            :value="type"
          >
            {{ type }}
          </option>
        </select>
      </label>
      <label>
        Search
        <input
          type="search"
          :value="props.querySearch"
          placeholder="Find by id, name, type, url..."
          @input="handleSearchChange"
        >
      </label>
      <button class="secondary" type="button" @click="emit('refresh')" :disabled="props.loading">
        Refresh list
      </button>
    </div>

    <p v-if="!props.loading && props.filteredQueries.length === 0" class="empty-message">
      No queries match the current filters.
    </p>

    <ul class="query-list">
      <li
        v-for="item in props.filteredQueries"
        :key="`${item.type}:${item.id}`"
        class="query-card"
      >
        <div class="query-header">
          <div class="query-title-group">
            <strong>{{ queryTitle(item) }}</strong>
            <span>{{ item.type }} / {{ item.id }}</span>
          </div>
          <div class="query-actions">
            <button class="secondary" type="button" @click="emit('edit', item)">
              Edit
            </button>
            <button class="secondary danger-btn" type="button" @click="emit('delete', item)">
              Delete
            </button>
          </div>
        </div>

        <div v-if="queryMeta(item).length" class="query-meta">
          <span v-for="entry in queryMeta(item)" :key="entry">{{ entry }}</span>
        </div>

        <details class="query-json">
          <summary>JSON</summary>
          <pre>{{ formatQuery(item) }}</pre>
        </details>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.query-toolbar {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(auto-fit, minmax(210px, 1fr));
  align-items: end;
}

.query-toolbar label {
  display: grid;
  gap: 5px;
  color: rgba(215, 222, 234, 0.68);
  font-size: 0.78rem;
  font-weight: 600;
}

.empty-message {
  margin: 14px 0 0;
  color: rgba(215, 222, 234, 0.66);
}

.list-head {
  margin-top: 22px;
}

.query-title-group {
  min-width: 0;
  display: grid;
  gap: 2px;
}

.query-title-group strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.query-title-group span {
  color: rgba(215, 222, 234, 0.52);
  font-size: 0.74rem;
}

.query-actions {
  display: flex;
  flex-shrink: 0;
  gap: 6px;
}

.query-actions button {
  padding: 5px 9px;
  font-size: 0.78rem;
}

.danger-btn:hover {
  border-color: rgba(255, 128, 128, 0.45);
  color: #ffb0b0;
}

.query-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.query-meta span {
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.035);
  border: 1px solid rgba(139, 152, 173, 0.12);
  color: rgba(215, 222, 234, 0.68);
  font-size: 0.72rem;
  padding: 2px 6px;
}

.query-json {
  margin-top: 8px;
  font-size: 0.76rem;
}

.query-json summary {
  color: rgba(215, 222, 234, 0.52);
  cursor: pointer;
  user-select: none;
}

.query-json pre {
  margin-top: 6px;
}

@media (max-width: 640px) {
  .query-header {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
