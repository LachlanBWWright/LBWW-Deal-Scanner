<script setup lang="ts">
import { computed, provide, ref, watch, type HTMLAttributes } from "vue";
import { cn } from "@/lib/utils";
import { tabsKey } from "./tabsContext";

defineOptions({
  name: "UiTabs",
});

const props = withDefaults(
  defineProps<{
    class?: HTMLAttributes["class"];
    modelValue?: string;
    defaultValue?: string;
  }>(),
  {
    modelValue: "",
    defaultValue: "",
  },
);

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
}>();

const currentValue = ref(props.modelValue || props.defaultValue);

watch(
  () => props.modelValue,
  (nextValue) => {
    currentValue.value = nextValue || props.defaultValue;
  },
);

function setValue(nextValue: string) {
  currentValue.value = nextValue;
  emit("update:modelValue", nextValue);
}

provide(tabsKey, {
  value: currentValue,
  setValue,
});

const classes = computed(() => cn(props.class));
</script>

<template>
  <div :class="classes">
    <slot />
  </div>
</template>
