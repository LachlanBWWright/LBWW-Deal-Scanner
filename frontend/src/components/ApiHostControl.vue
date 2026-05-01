<script setup lang="ts">
import { ref, watch } from "vue";

const props = defineProps<{
  apiHost: string;
  apiSecret: string;
}>();

const emit = defineEmits<{
  (e: "update:apiHost", value: string): void;
  (e: "update:apiSecret", value: string): void;
  (e: "save"): void;
  (e: "reset"): void;
}>();

const localApiHost = ref(props.apiHost);
const localApiSecret = ref(props.apiSecret);

watch(
  () => props.apiHost,
  (value) => {
    localApiHost.value = value;
  },
);

watch(
  () => props.apiSecret,
  (value) => {
    localApiSecret.value = value;
  },
);

function updateHost() {
  emit("update:apiHost", localApiHost.value);
}

function updateSecret() {
  emit("update:apiSecret", localApiSecret.value);
}

function saveHost() {
  updateHost();
  updateSecret();
  emit("save");
}

function resetHost() {
  emit("reset");
}
</script>
<script lang="ts">
export default {};
</script>

<template>
  <div class="field-grid">
    <div class="field-group">
      <label for="api-host">API host</label>
      <input
        id="api-host"
        type="text"
        placeholder="Leave empty for current host"
        v-model="localApiHost"
      />
    </div>
    <div class="field-group">
      <label for="api-secret">API secret</label>
      <input
        id="api-secret"
        type="password"
        placeholder="Enter API secret"
        v-model="localApiSecret"
      />
    </div>
    <div class="button-row">
      <button class="primary" type="button" @click="saveHost">
        Save API settings
      </button>
      <button class="secondary" type="button" @click="resetHost">
        Reset host
      </button>
    </div>
  </div>
</template>
