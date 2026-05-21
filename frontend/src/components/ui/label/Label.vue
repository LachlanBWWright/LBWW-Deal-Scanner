<script setup lang="ts">
import { computed, type HTMLAttributes } from "vue";
import { cn } from "@/lib/utils";

defineOptions({
  name: "UiLabel",
});

const props = defineProps<{
  class?: HTMLAttributes["class"];
}>();

const classes = computed(() => cn("text-sm font-medium leading-none text-foreground", props.class));
</script>

<template>
  <label :class="classes">
    <slot />
  </label>
</template>
