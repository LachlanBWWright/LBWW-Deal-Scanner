<script setup lang="ts">
import { computed, type HTMLAttributes } from "vue";
import { cn } from "@/lib/utils";

defineOptions({
  name: "UiTabsList",
});

const props = defineProps<{
  class?: HTMLAttributes["class"];
}>();

const classes = computed(() =>
  cn(
    "inline-flex h-10 items-center justify-start rounded-md bg-muted p-1 text-muted-foreground",
    props.class,
  ),
);
</script>

<template>
  <div role="tablist" :class="classes">
    <slot />
  </div>
</template>
