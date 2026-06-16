<script setup lang="ts">
import { computed, inject, type HTMLAttributes } from "vue";
import { cn } from "@/lib/utils";
import { tabsKey } from "./tabsContext";

defineOptions({
  name: "UiTabsContent",
});

const props = defineProps<{
  class?: HTMLAttributes["class"];
  value: string;
}>();

const tabs = inject(tabsKey, null);
const isActive = computed(() => tabs?.value.value === props.value);

const classes = computed(() => cn(props.class));
</script>

<template>
  <div v-if="isActive" :class="classes">
    <slot />
  </div>
</template>
