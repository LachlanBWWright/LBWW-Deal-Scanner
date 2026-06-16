<script setup lang="ts">
import { Button } from "@/components/ui/button";
import type { QueryType } from "./queryTypes";
import { formatQueryTypeLabel } from "./queryTypes";

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
  <div class="space-y-2">
    <p class="text-sm font-medium text-foreground">Query type</p>
    <div class="flex flex-wrap gap-2">
      <Button
        v-for="type in props.supportedTypes"
        :key="type"
        type="button"
        size="sm"
        :variant="props.selectedType === type ? 'default' : 'outline'"
        @click="chooseType(type)"
      >
        {{ formatQueryTypeLabel(type) }}
      </Button>
    </div>
  </div>
</template>
