<script setup lang="ts">
import { useQueryManager } from "../composables/useQueryManager";
import type { QueryFormState, QueryType } from "./queryTypes";
import ApiHostControl from "./ApiHostControl.vue";
import QueryTypePicker from "./QueryTypePicker.vue";
import QueryEditor from "./QueryEditor.vue";
import QueryList from "./QueryList.vue";

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
  formFields,
  setApiHost,
  clearApiHost,
  selectType,
  saveQuery,
  resetForm,
  refreshQueries,
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
  <section class="panel query-manager">
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
  </section>
</template>

<style>
.field-grid {
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  margin-top: 18px;
}

.field-group {
  display: grid;
  gap: 8px;
}

.field-group label,
.label {
  color: rgba(219, 227, 240, 0.84);
  font-size: 0.95rem;
}

input,
select {
  width: 100%;
  min-height: 44px;
  border-radius: 18px;
  border: 1px solid rgba(154, 173, 201, 0.18);
  background: rgba(255, 255, 255, 0.04);
  color: inherit;
  padding: 0 14px;
}

input:focus,
select:focus {
  outline: 2px solid rgba(125, 253, 212, 0.5);
}

.type-selector {
  grid-column: 1 / -1;
}

.type-options {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.type-button {
  border-radius: 999px;
  border: 1px solid rgba(154, 173, 201, 0.22);
  background: rgba(255, 255, 255, 0.05);
  color: inherit;
  padding: 12px 16px;
  cursor: pointer;
}

.type-button.active {
  background: linear-gradient(135deg, #7dfdd4 0%, #3ea7ff 100%);
  color: #04111f;
}

.field-switch {
  align-items: center;
  display: inline-flex;
  gap: 10px;
  color: rgba(219, 227, 240, 0.86);
  font-size: 0.95rem;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 18px;
}

button {
  min-width: 140px;
}

.success-message {
  margin-top: 14px;
  color: #67d8ff;
}

.query-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
}

.panel-head.compact {
  margin-bottom: 12px;
}

.query-stats {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  margin-top: 18px;
}

.query-stat-card {
  display: grid;
  gap: 2px;
  border-radius: 16px;
  padding: 12px;
  border: 1px solid rgba(154, 173, 201, 0.12);
  background: rgba(255, 255, 255, 0.04);
}

.query-stat-card strong {
  font-size: 1.3rem;
  line-height: 1;
}

.query-stat-card span {
  color: rgba(219, 227, 240, 0.86);
  font-size: 0.86rem;
}

.query-review {
  margin-top: 18px;
  border-radius: 18px;
  border: 1px solid rgba(154, 173, 201, 0.12);
  background: rgba(255, 255, 255, 0.02);
  padding: 14px;
}

.query-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 18px;
  margin-top: 18px;
}

.query-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(154, 173, 201, 0.14);
  border-radius: 20px;
  padding: 18px;
}

.query-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  margin-bottom: 12px;
  color: rgba(231, 236, 244, 0.92);
}

pre {
  margin: 0;
  padding: 14px;
  border-radius: 16px;
  background: rgba(12, 20, 35, 0.9);
  overflow-x: auto;
  font-size: 0.92rem;
  color: rgba(219, 227, 240, 0.86);
}

.button-row {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 12px;
}
</style>
