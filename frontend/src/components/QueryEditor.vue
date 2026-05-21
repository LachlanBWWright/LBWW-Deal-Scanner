<script setup lang="ts">
import { computed } from "vue";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { QueryItem, QueryType, QueryFormState, FormField } from "./queryTypes";
import { formatQueryTypeLabel } from "./queryTypes";

const props = defineProps<{
  selectedType: QueryType;
  form: QueryFormState;
  formFields: FormField[];
  editingQuery: QueryItem | null;
  saving: boolean;
}>();

const emit = defineEmits<{
  (e: "update:form", value: QueryFormState): void;
  (e: "save"): void;
  (e: "reset"): void;
}>();

const isEditing = computed(() => props.editingQuery !== null);
const title = computed(() => (isEditing.value ? "Update query" : "Create query"));
const subtitle = computed(() => {
  if (!props.editingQuery) {
    return "Create a new saved query with the selected source rules.";
  }

  const label = formatQueryTypeLabel(props.editingQuery.type);
  const identity =
    props.editingQuery.name ||
    props.editingQuery.displayUrl ||
    props.editingQuery.url ||
    props.editingQuery.id;
  return `Editing ${label} · ${identity}`;
});

function updateField(field: FormField, rawValue: string) {
  const value =
    field.type === "number" && rawValue !== "" ? Number(rawValue) : rawValue;
  emit("update:form", {
    ...props.form,
    [field.key]: value,
  });
}

function updateDmOnly(value: boolean) {
  emit("update:form", {
    ...props.form,
    dmOnly: value,
  });
}

function onFieldInput(field: FormField, event: Event) {
  const target = event.target;
  if (!(target instanceof HTMLInputElement)) return;
  updateField(field, target.value);
}

function fieldValue(key: FormField["key"]): string | number {
  const value = props.form[key];
  return typeof value === "boolean" ? Number(value) : value;
}
</script>
<script lang="ts">
export default {};
</script>

<template>
  <form class="space-y-4" @submit.prevent="$emit('save')">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <p class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">
          {{ isEditing ? "Edit mode" : "New query" }}
        </p>
        <h3 class="text-lg font-semibold tracking-tight">{{ title }}</h3>
        <p class="text-sm text-muted-foreground">
          {{ subtitle }}
        </p>
      </div>
      <p v-if="isEditing" class="rounded-full border border-sky-200 bg-sky-100 px-3 py-1 text-xs font-semibold text-sky-800">Editing</p>
    </div>

    <div class="grid gap-3 sm:grid-cols-2">
      <div
        v-for="field in props.formFields"
        :key="field.key"
        class="space-y-2"
      >
        <Label :for="field.key">{{ field.label }}</Label>
        <Input
          :id="field.key"
          :type="field.type"
          :value="fieldValue(field.key)"
          :placeholder="field.key === 'displayUrl' ? 'Optional' : ''"
          @input="onFieldInput(field, $event)"
        />
      </div>
    </div>

    <div class="flex items-center gap-2 rounded-md border border-border/60 bg-muted/40 px-3 py-2">
      <Checkbox :model-value="props.form.dmOnly" @update:model-value="updateDmOnly" />
      <Label>Send notifications as DM only</Label>
    </div>

    <div class="flex flex-wrap gap-2">
      <Button :disabled="props.saving" type="submit">
        {{ title }}
      </Button>
      <Button variant="secondary" type="button" @click="$emit('reset')" :disabled="props.saving">
        {{ isEditing ? "Cancel edit" : "Reset form" }}
      </Button>
    </div>
  </form>
</template>
