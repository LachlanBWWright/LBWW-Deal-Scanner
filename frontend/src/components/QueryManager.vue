<script setup lang="ts">
import { useQueryManager } from "../composables/useQueryManager";
import type { QueryFormState, QueryType } from "./queryTypes";
import ApiHostControl from "./ApiHostControl.vue";
import QueryTypePicker from "./QueryTypePicker.vue";
import QueryEditor from "./QueryEditor.vue";
import QueryList from "./QueryList.vue";
import QueryResultsList from "./QueryResultsList.vue";

const {
  supportedQueryTypes,
  apiHost,
  apiSecret,
  querySearch,
  queryStats,
  selectedQueryPreview,
  lastSyncedAt,
  filterType,
  selectedType,
  editingQuery,
  loading,
  saving,
  errorMessage,
  successMessage,
  form,
  filteredQueries,
  searchResults,
  formFields,
  setApiHost,
  clearApiHost,
  selectType,
  saveQuery,
  resetForm,
  refreshQueries,
  refreshSearchResults,
  startEditing,
  deleteQuery,
} = useQueryManager();

function updateApiHost(value: string) {
  apiHost.value = value;
}

function updateApiSecret(value: string) {
  apiSecret.value = value;
}

function updateForm(value: QueryFormState) {
  form.value = value;
}

function updateFilterType(value: QueryType | "all") {
  filterType.value = value;
}

function updateQuerySearch(value: string) {
  querySearch.value = value;
}

function formatQueryPreview() {
  if (!selectedQueryPreview.value) return "No query selected yet.";
  return JSON.stringify(selectedQueryPreview.value, null, 2);
}

function syncedAgoLabel() {
  if (!lastSyncedAt.value) return "Not synced yet";
  return `Last synced ${new Date(lastSyncedAt.value).toLocaleTimeString()}`;
}
</script>
<script lang="ts">
export default {};
</script>

<template>
  <section class="query-manager">
    <div class="panel-head">
      <h2>Query manager</h2>
      <span>Manage saved watch queries and API settings.</span>
    </div>

    <ApiHostControl
      :apiHost="apiHost"
      :apiSecret="apiSecret"
      @update:apiHost="updateApiHost"
      @update:apiSecret="updateApiSecret"
      @save="setApiHost"
      @reset="clearApiHost"
    />

    <QueryTypePicker
      :supportedTypes="supportedQueryTypes"
      :selectedType="selectedType"
      @update:selectedType="selectType"
    />

    <QueryEditor
      :selectedType="selectedType"
      :form="form"
      :formFields="formFields"
      :editingQuery="!!editingQuery"
      :saving="saving"
      @update:form="updateForm"
      @save="saveQuery"
      @reset="resetForm"
    />

    <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
    <p v-if="successMessage" class="success-message">{{ successMessage }}</p>

    <div class="query-stats">
      <article v-for="entry in queryStats" :key="entry.type" class="query-stat-card">
        <strong>{{ entry.count }}</strong>
        <span>{{ entry.type }}</span>
      </article>
    </div>

    <div class="query-review">
      <div class="panel-head compact">
        <h2>Query preview</h2>
        <span>{{ syncedAgoLabel() }}</span>
      </div>
      <pre>{{ formatQueryPreview() }}</pre>
    </div>

    <QueryList
      :filteredQueries="filteredQueries"
      :supportedTypes="supportedQueryTypes"
      :filterType="filterType"
      :querySearch="querySearch"
      :loading="loading"
      @update:filterType="updateFilterType"
      @update:querySearch="updateQuerySearch"
      @refresh="refreshQueries"
      @edit="startEditing"
      @delete="deleteQuery"
    />

    <QueryResultsList
      :searchResults="searchResults"
      :loading="loading"
      @refresh="refreshSearchResults"
    />
  </section>
</template>

<style>
.query-manager {
  max-width: 1180px;
  margin: 0 auto;
  padding: 16px;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: center;
  margin-bottom: 10px;
}

.panel-head h2 {
  margin: 0;
  font-size: 0.96rem;
  letter-spacing: 0;
}

.panel-head span {
  color: rgba(215, 222, 234, 0.58);
  font-size: 0.78rem;
}

.field-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
  margin-top: 12px;
}

.field-group {
  display: grid;
  gap: 5px;
}

.field-group label,
.label {
  color: rgba(215, 222, 234, 0.68);
  font-size: 0.78rem;
  font-weight: 600;
}

input,
select {
  width: 100%;
  min-height: 34px;
  border-radius: 6px;
  border: 1px solid rgba(139, 152, 173, 0.2);
  background: rgba(255, 255, 255, 0.04);
  color: inherit;
  padding: 0 10px;
  font: inherit;
  font-size: 0.86rem;
}

input:focus,
select:focus {
  outline: 2px solid rgba(87, 166, 255, 0.45);
  outline-offset: 1px;
}

.type-selector {
  grid-column: 1 / -1;
}

.type-options {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.type-button {
  min-width: 0;
  border-radius: 6px;
  border: 1px solid rgba(139, 152, 173, 0.22);
  background: rgba(255, 255, 255, 0.035);
  color: inherit;
  padding: 7px 10px;
  cursor: pointer;
  font-size: 0.82rem;
}

.type-button.active {
  background: rgba(47, 129, 247, 0.18);
  border-color: rgba(87, 166, 255, 0.55);
  color: #eef3fb;
}

.field-switch {
  align-items: center;
  display: inline-flex;
  gap: 8px;
  color: rgba(215, 222, 234, 0.78);
  font-size: 0.85rem;
  margin-top: 10px;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

button {
  min-width: 0;
}

.success-message {
  margin-top: 14px;
  color: #67d8ff;
}

.query-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
}

.panel-head.compact {
  margin-bottom: 12px;
}

.query-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.query-stat-card {
  display: flex;
  align-items: center;
  gap: 8px;
  border-radius: 6px;
  padding: 5px 9px;
  border: 1px solid rgba(139, 152, 173, 0.12);
  background: rgba(255, 255, 255, 0.025);
}

.query-stat-card strong {
  font-size: 1rem;
  line-height: 1;
}

.query-stat-card span {
  color: rgba(219, 227, 240, 0.7);
  font-size: 0.78rem;
}

.query-review {
  margin-top: 14px;
  border-radius: 8px;
  border: 1px solid rgba(139, 152, 173, 0.12);
  background: rgba(255, 255, 255, 0.018);
  padding: 10px;
}

.query-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 8px;
  margin-top: 12px;
}

.query-card {
  background: rgba(255, 255, 255, 0.025);
  border: 1px solid rgba(139, 152, 173, 0.12);
  border-radius: 8px;
  padding: 10px;
}

.query-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  margin-bottom: 8px;
  color: rgba(238, 243, 251, 0.92);
}

pre {
  margin: 0;
  padding: 10px;
  border-radius: 6px;
  background: rgba(7, 12, 20, 0.82);
  overflow-x: auto;
  font-size: 0.78rem;
  color: rgba(215, 222, 234, 0.78);
}

.button-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 10px;
}
</style>
