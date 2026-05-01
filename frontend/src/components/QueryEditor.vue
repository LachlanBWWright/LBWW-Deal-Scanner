<script setup lang="ts">
import { computed, ref, watch } from "vue";
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

const localForm = ref<QueryFormState>({ ...props.form });

watch(
  () => props.form,
  (value) => {
    localForm.value = { ...value };
  },
  { deep: true },
);

watch(
  localForm,
  (value) => {
    emit("update:form", value);
  },
  { deep: true },
);

const title = computed(() => (props.editingQuery ? "Update query" : "Create query"));
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
          v-model="localForm[field.key]"
          :placeholder="field.key === 'displayUrl' ? 'Optional' : ''"
        />
      </div>
    </div>

    <div class="field-group">
      <label class="field-switch">
        <input type="checkbox" v-model="localForm.dmOnly" />
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
