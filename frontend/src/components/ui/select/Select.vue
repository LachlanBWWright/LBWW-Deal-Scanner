<script setup lang="ts">
import { computed, type HTMLAttributes, useAttrs } from "vue";
import { cn } from "@/lib/utils";

defineOptions({
  name: "UiSelect",
});

const props = defineProps<{
  class?: HTMLAttributes["class"];
  modelValue?: string;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
}>();

const attrs = useAttrs();
const classes = computed(() =>
  cn(
    "flex h-10 w-full rounded-md border border-input bg-card px-3 py-2 text-sm text-foreground shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
    props.class,
  ),
);

function onChange(event: Event) {
  const target = event.target;
  if (!(target instanceof HTMLSelectElement)) return;
  emit("update:modelValue", target.value);
}
</script>

<template>
  <select v-bind="attrs" :value="props.modelValue" :class="classes" @change="onChange">
    <slot />
  </select>
</template>
