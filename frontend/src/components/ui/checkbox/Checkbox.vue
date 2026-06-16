<script setup lang="ts">
import { computed, type HTMLAttributes, useAttrs } from "vue";
import { cn } from "@/lib/utils";

defineOptions({
  name: "UiCheckbox",
});

const props = defineProps<{
  class?: HTMLAttributes["class"];
  modelValue?: boolean;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
}>();

const attrs = useAttrs();
const classes = computed(() =>
  cn(
    "h-4 w-4 rounded border border-input text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
    props.class,
  ),
);

function onChange(event: Event) {
  const target = event.target;
  if (!(target instanceof HTMLInputElement)) return;
  emit("update:modelValue", target.checked);
}
</script>

<template>
  <input
    v-bind="attrs"
    type="checkbox"
    :checked="Boolean(props.modelValue)"
    :class="classes"
    @change="onChange"
  />
</template>
