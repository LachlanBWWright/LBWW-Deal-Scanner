<script setup lang="ts">
import { computed, type HTMLAttributes } from "vue";
import { cn } from "@/lib/utils";

defineOptions({
  name: "UiCard",
});

const props = defineProps<{ class?: HTMLAttributes["class"] }>();
const classes = computed(() => cn("rounded-xl border bg-card text-card-foreground shadow-sm", props.class));
</script>

<template>
  <section :class="classes">
    <slot />
  </section>
</template>
