<script setup lang="ts">
import type { QueryType } from "./queryTypes";

const props = defineProps<{
  supportedTypes: QueryType[];
  selectedType: QueryType;
}>();

const emit = defineEmits<{
  (e: "update:selectedType", value: QueryType): void;
}>();

function chooseType(type: QueryType) {
  emit("update:selectedType", type);
}
</script>
<script lang="ts">
export default {};
</script>

<template>
  <div class="field-grid">
    <div class="field-group type-selector">
      <span class="label">Query type</span>
      <div class="type-options">
        <button
          v-for="type in props.supportedTypes"
          :key="type"
          type="button"
          :class="[{ active: props.selectedType === type }, 'type-button']"
          @click="chooseType(type)"
        >
          {{ type }}
        </button>
      </div>
    </div>
  </div>
</template>
