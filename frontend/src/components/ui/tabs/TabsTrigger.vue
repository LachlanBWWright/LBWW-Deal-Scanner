<script setup lang="ts">
import { computed, inject, type HTMLAttributes } from "vue";
import { cn } from "@/lib/utils";
import { tabsKey } from "./tabsContext";

defineOptions({
  name: "UiTabsTrigger",
});

const props = defineProps<{
  class?: HTMLAttributes["class"];
  value: string;
}>();

const tabs = inject(tabsKey, null);

const isActive = computed(() => tabs?.value.value === props.value);

const classes = computed(() =>
  cn(
    "inline-flex items-center justify-center rounded-sm px-3 py-1.5 text-sm font-medium transition-all",
    isActive.value
      ? "bg-background text-foreground shadow-sm"
      : "text-muted-foreground hover:text-foreground",
    props.class,
  ),
);

function onClick() {
  tabs?.setValue(props.value);
}
</script>

<template>
  <button
    type="button"
    role="tab"
    :aria-selected="isActive"
    :tabindex="isActive ? 0 : -1"
    :class="classes"
    :data-state="isActive ? 'active' : 'inactive'"
    @click="onClick"
  >
    <slot />
  </button>
</template>
