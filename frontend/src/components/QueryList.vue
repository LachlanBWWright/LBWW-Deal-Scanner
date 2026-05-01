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
</script>
<script lang="ts">
export default {};
</script>

<template>
  <div>
    <div class="panel-head" style="margin-top: 32px;">
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
          <strong>{{ item.type }}</strong>
          <span>{{ item.id }}</span>
        </div>

        <pre>{{ formatQuery(item) }}</pre>

        <div class="button-row">
          <button class="secondary" type="button" @click="emit('edit', item)">
            Edit
          </button>
          <button class="secondary" type="button" @click="emit('delete', item)">
            Delete
          </button>
        </div>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.query-toolbar {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(210px, 1fr));
  align-items: end;
}

.query-toolbar label {
  display: grid;
  gap: 6px;
}

.empty-message {
  margin: 14px 0 0;
  color: rgba(219, 227, 240, 0.78);
}
</style>
