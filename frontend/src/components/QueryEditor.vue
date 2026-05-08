<script setup lang="ts">
import { computed } from "vue";
import type { QueryType, QueryFormState, FormField } from "./queryTypes";

const props = defineProps<{
  selectedType: QueryType;
  form: QueryFormState;
  formFields: FormField[];
  editingQuery: boolean;
  saving: boolean;
}>();

const emit = defineEmits<{
  (e: "update:form", value: QueryFormState): void;
  (e: "save"): void;
  (e: "reset"): void;
}>();

const title = computed(() => (props.editingQuery ? "Update query" : "Create query"));

function updateField(field: FormField, rawValue: string) {
  const value = field.type === "number" ? Number(rawValue) : rawValue;
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
  const target = event.target as HTMLInputElement | null;
  if (!target) return;
  updateField(field, target.value);
}

function onDmOnlyChange(event: Event) {
  const target = event.target as HTMLInputElement | null;
  if (!target) return;
  updateDmOnly(target.checked);
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
  <form @submit.prevent="$emit('save')">
    <p class="editor-note">
      {{ props.editingQuery ? "Editing existing query values." : "Create a new query using the selected type." }}
    </p>
    <div class="field-grid">
      <div
        v-for="field in props.formFields"
        :key="field.key"
        class="field-group"
      >
        <label :for="field.key">{{ field.label }}</label>
        <input
          :id="field.key"
          :type="field.type"
          :value="fieldValue(field.key)"
          :placeholder="field.key === 'displayUrl' ? 'Optional' : ''"
          @input="onFieldInput(field, $event)"
        />
      </div>
    </div>

    <div class="field-group">
      <label class="field-switch">
        <input
          type="checkbox"
          :checked="props.form.dmOnly"
          @change="onDmOnlyChange($event)"
        />
        Send notifications as DM only
      </label>
    </div>

    <div class="hero-actions">
      <button class="primary" :disabled="props.saving" type="submit">
        {{ title }}
      </button>
      <button class="secondary" type="button" @click="$emit('reset')" :disabled="props.saving">
        Reset form
      </button>
    </div>
  </form>
</template>

<style scoped>
.editor-note {
  margin: 0 0 10px;
  color: rgba(219, 227, 240, 0.78);
}
</style>
