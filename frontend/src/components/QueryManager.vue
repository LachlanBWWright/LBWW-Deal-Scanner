<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { useQueryManager } from "../composables/useQueryManager";
import type { QueryFormState, QueryItem, QueryType } from "./queryTypes";
import { formatQueryTypeLabel } from "./queryTypes";
import QueryTypePicker from "./QueryTypePicker.vue";
import QueryEditor from "./QueryEditor.vue";
import QueryList from "./QueryList.vue";

const props = defineProps<{
  apiHost: string;
  apiSecret: string;
}>();

const editorPanelRef = ref<HTMLElement | null>(null);
const editorOpen = ref(false);
const pendingDeleteQuery = ref<QueryItem | null>(null);

const {
  supportedQueryTypes,
  querySearch,
  queryStats,
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
  selectType,
  saveQuery,
  resetForm,
  refreshQueries,
  startEditing,
  deleteQuery,
} = useQueryManager(
  computed(() => props.apiHost),
  computed(() => props.apiSecret),
);

function updateForm(value: QueryFormState) {
  form.value = value;
}

function updateFilterType(value: QueryType | "all") {
  filterType.value = value;
}

function updateQuerySearch(value: string) {
  querySearch.value = value;
}

function isQueryType(value: string): value is QueryType {
  return supportedQueryTypes.some((type) => type === value);
}

function handleSearchInput(event: Event) {
  const target = event.target;
  if (!(target instanceof HTMLInputElement)) return;
  updateQuerySearch(target.value);
}

function handleFilterInput(event: Event) {
  const target = event.target;
  if (!(target instanceof HTMLSelectElement)) return;
  if (target.value === "all") {
    updateFilterType("all");
    return;
  }
  if (isQueryType(target.value)) {
    updateFilterType(target.value);
  }
}

function openNewQuery() {
  resetForm();
  editorOpen.value = true;
}

function openExistingQuery(item: QueryItem) {
  startEditing(item);
  editorOpen.value = true;
}

function closeEditor() {
  editorOpen.value = false;
  resetForm();
}

async function submitEditor() {
  const saved = await saveQuery();
  if (saved) {
    closeEditor();
  }
}

function requestDelete(item: QueryItem) {
  pendingDeleteQuery.value = item;
}

function cancelDelete() {
  pendingDeleteQuery.value = null;
}

async function confirmDelete() {
  if (!pendingDeleteQuery.value) return;
  const deleted = await deleteQuery(pendingDeleteQuery.value);
  if (deleted) {
    pendingDeleteQuery.value = null;
  }
}

watch(editingQuery, async (value, previous) => {
  if (!value || value === previous) return;
  editorOpen.value = true;
  await nextTick();
  editorPanelRef.value?.scrollIntoView({
    behavior: "smooth",
    block: "start",
  });
});
</script>
<script lang="ts">
export default {};
</script>

<template>
  <section class="mx-auto max-w-6xl space-y-4">
    <Card>
      <CardHeader class="gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Query manager</p>
          <CardTitle>Saved queries</CardTitle>
          <CardDescription>Search first. Open the editor only when you need it.</CardDescription>
        </div>
        <div class="flex flex-wrap gap-2">
          <Badge v-for="entry in queryStats" :key="entry.type" variant="secondary" class="gap-1">
            <strong>{{ entry.count }}</strong>
            <span>{{ entry.typeLabel }}</span>
          </Badge>
        </div>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="grid gap-3 md:grid-cols-[minmax(0,1.6fr)_minmax(220px,0.8fr)_auto] md:items-end">
          <div class="space-y-2">
            <Label>Search queries</Label>
            <Input
              :value="querySearch"
              type="search"
              placeholder="Search by name, type, url, id..."
              @input="handleSearchInput"
            />
          </div>

          <div class="space-y-2">
            <Label>Filter type</Label>
            <Select :value="filterType" @change="handleFilterInput">
              <option value="all">All types</option>
              <option
                v-for="type in supportedQueryTypes"
                :key="type"
                :value="type"
              >
                {{ formatQueryTypeLabel(type) }}
              </option>
            </Select>
          </div>

          <div class="flex items-end gap-2">
            <Button variant="secondary" type="button" @click="refreshQueries" :disabled="loading">
              Refresh
            </Button>
            <Button type="button" title="New query" @click="openNewQuery">
              + New
            </Button>
          </div>
        </div>

        <p v-if="errorMessage" class="rounded-md border border-red-500/40 bg-red-500/10 px-3 py-2 text-sm font-medium text-red-200">{{ errorMessage }}</p>
        <p v-if="successMessage" class="rounded-md border border-emerald-500/40 bg-emerald-500/10 px-3 py-2 text-sm font-medium text-emerald-200">{{ successMessage }}</p>

        <QueryList
          :filteredQueries="filteredQueries"
          :searchResults="searchResults"
          :editingQueryId="editingQuery?.id ?? null"
          :editingQueryType="editingQuery?.type ?? null"
          @edit="openExistingQuery"
          @delete="requestDelete"
        />
      </CardContent>
    </Card>

    <teleport to="body">
      <div v-if="editorOpen" class="fixed inset-0 z-40 flex items-center justify-center bg-slate-900/70 p-4 backdrop-blur-sm" @click.self="closeEditor">
        <section ref="editorPanelRef" class="max-h-[86vh] w-full max-w-3xl overflow-auto rounded-2xl border border-border bg-card p-5 shadow-xl" role="dialog" aria-modal="true">
          <div class="mb-4 flex items-start justify-between gap-3">
            <div>
              <p class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">{{ editingQuery ? "Edit query" : "New query" }}</p>
              <h3 class="text-lg font-semibold tracking-tight">{{ editingQuery ? "Update saved query" : "Create saved query" }}</h3>
              <p class="text-sm text-muted-foreground">{{ editingQuery ? "Change the rule, then save the updates." : "Pick a source type and define the query." }}</p>
            </div>
            <Button variant="ghost" type="button" title="Close editor" @click="closeEditor">✕</Button>
          </div>

          <div v-if="editingQuery" class="mb-4 rounded-md border border-border/70 bg-muted/40 px-3 py-2">
            <p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">Query type</p>
            <p class="text-sm font-semibold">{{ formatQueryTypeLabel(editingQuery.type) }}</p>
          </div>
          <QueryTypePicker
            v-else
            :supportedTypes="supportedQueryTypes"
            :selectedType="selectedType"
            @update:selectedType="selectType"
          />

          <div class="mt-4">
            <QueryEditor
              :selectedType="selectedType"
              :form="form"
              :formFields="formFields"
              :editingQuery="editingQuery"
              :saving="saving"
              @update:form="updateForm"
              @save="submitEditor"
              @reset="resetForm"
            />
          </div>
        </section>
      </div>
    </teleport>

    <teleport to="body">
      <div v-if="pendingDeleteQuery" class="fixed inset-0 z-40 flex items-center justify-center bg-slate-900/70 p-4 backdrop-blur-sm" @click.self="cancelDelete">
        <section class="w-full max-w-lg rounded-2xl border border-border bg-card p-5 shadow-xl" role="dialog" aria-modal="true">
          <p class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Delete query</p>
          <h3 class="mt-1 text-lg font-semibold">Delete this saved query?</h3>
          <p class="mt-2 break-all text-sm text-foreground">
            {{ pendingDeleteQuery.name || pendingDeleteQuery.displayUrl || pendingDeleteQuery.url || pendingDeleteQuery.id }}
          </p>
          <p class="mt-2 text-sm text-muted-foreground">
            This removes the saved rule and cannot be undone.
          </p>
          <div class="mt-4 flex justify-end gap-2">
            <Button variant="secondary" type="button" @click="cancelDelete" :disabled="saving">Cancel</Button>
            <Button variant="destructive" type="button" @click="confirmDelete" :disabled="saving">
              Delete query
            </Button>
          </div>
        </section>
      </div>
    </teleport>
  </section>
</template>
