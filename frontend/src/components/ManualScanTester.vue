<script setup lang="ts">
import { ref, computed } from "vue";
import { useTestingConsole } from "../composables/useTestingConsole";

const {
  capabilities,
  lastScanResult,
  scanning,
  errorMessage,
  successMessage,
  runTestScan,
} = useTestingConsole();

const supportedScannerTypes = ["cashConverters", "ebay", "gumtree", "salvos"] as const;
type SupportedType = (typeof supportedScannerTypes)[number];

const scannerType = ref<SupportedType>("ebay");
const notifyToggle = ref(false);

// Per-scanner payload fields
const url = ref("");
const maxPrice = ref<number | "">("");
const minPrice = ref<number | "">(0);
const name = ref("");
const requiredPhrases = ref("");
const excludePhrases = ref("");

const testingEnabled = computed(() => capabilities.value?.testingEnabled ?? false);

const payloadPreview = computed(() => {
  return JSON.stringify(buildPayload(), null, 2);
});

function buildPayload(): Record<string, unknown> {
  switch (scannerType.value) {
    case "ebay":
    case "gumtree":
      return { url: url.value, maxPrice: maxPrice.value };
    case "cashConverters":
      return {
        url: url.value,
        requiredPhrases: requiredPhrases.value,
        excludePhrases: excludePhrases.value,
      };
    case "salvos":
      return {
        name: name.value,
        minPrice: minPrice.value,
        maxPrice: maxPrice.value,
      };
    default:
      return {};
  }
}

function resetFields() {
  url.value = "";
  maxPrice.value = "";
  minPrice.value = 0;
  name.value = "";
  requiredPhrases.value = "";
  excludePhrases.value = "";
}

async function run() {
  if (!testingEnabled.value) return;

  await runTestScan({
    type: scannerType.value,
    payload: buildPayload(),
    notify: notifyToggle.value,
  });
}
</script>
<script lang="ts">
export default {};
</script>

<template>
  <div class="tester-panel">
    <div class="tester-header">
      <h3>Manual Scan Tester</h3>
      <span class="hint">Temporary scan input — no user query is saved</span>
      <div v-if="!testingEnabled" class="badge badge-disabled">Testing disabled</div>
      <div v-else class="badge badge-enabled">Testing enabled</div>
    </div>

    <div class="field-grid">
      <label class="field">
        <span>Scanner type</span>
        <select v-model="scannerType" @change="resetFields">
          <option v-for="t in supportedScannerTypes" :key="t" :value="t">{{ t }}</option>
        </select>
      </label>

      <label class="field checkbox-field">
        <input v-model="notifyToggle" type="checkbox" />
        <span>Publish notifications</span>
      </label>

      <!-- eBay / Gumtree fields -->
      <template v-if="scannerType === 'ebay' || scannerType === 'gumtree'">
        <label class="field field-wide">
          <span>Query URL</span>
          <input v-model="url" type="url" placeholder="https://..." />
        </label>
        <label class="field">
          <span>Max price</span>
          <input v-model.number="maxPrice" type="number" step="0.01" min="0" />
        </label>
      </template>

      <!-- Cash Converters fields -->
      <template v-if="scannerType === 'cashConverters'">
        <label class="field field-wide">
          <span>Query URL</span>
          <input v-model="url" type="url" placeholder="https://..." />
        </label>
        <label class="field">
          <span>Required phrases</span>
          <input v-model="requiredPhrases" type="text" placeholder="(optional)" />
        </label>
        <label class="field">
          <span>Exclude phrases</span>
          <input v-model="excludePhrases" type="text" placeholder="(optional)" />
        </label>
      </template>

      <!-- Salvos fields -->
      <template v-if="scannerType === 'salvos'">
        <label class="field">
          <span>Item name</span>
          <input v-model="name" type="text" placeholder="e.g. iPhone" />
        </label>
        <label class="field">
          <span>Min price</span>
          <input v-model.number="minPrice" type="number" step="0.01" min="0" />
        </label>
        <label class="field">
          <span>Max price</span>
          <input v-model.number="maxPrice" type="number" step="0.01" min="0" />
        </label>
      </template>
    </div>

    <details class="payload-preview">
      <summary>Payload preview</summary>
      <pre>{{ payloadPreview }}</pre>
    </details>

    <div class="tester-actions">
      <button
        class="primary run-btn"
        :disabled="scanning || !testingEnabled"
        @click="run"
      >
        {{ scanning ? "Running scan..." : "Run scan" }}
      </button>
    </div>

    <p v-if="errorMessage" class="msg msg-error">{{ errorMessage }}</p>
    <p v-if="successMessage" class="msg msg-success">{{ successMessage }}</p>

    <div v-if="lastScanResult" class="result-box">
      <p class="result-label">Scan result</p>

      <div class="result-row">
        <span class="result-key">Duration</span>
        <span>{{ lastScanResult.durationMs }}ms</span>
      </div>
      <div class="result-row">
        <span class="result-key">Published</span>
        <span>{{ lastScanResult.notificationsPublished ? "Yes" : "No" }}</span>
      </div>

      <template v-if="lastScanResult.errors.length">
        <p class="section-label">Errors</p>
        <div
          v-for="(err, idx) in lastScanResult.errors"
          :key="idx"
          class="result-row"
        >
          <span class="outcome-failed">{{ err }}</span>
        </div>
      </template>

      <template v-if="lastScanResult.notifications.length">
        <p class="section-label">Notifications ({{ lastScanResult.notifications.length }})</p>
        <div
          v-for="notif in lastScanResult.notifications"
          :key="notif.url"
          class="notif-card"
        >
          <div class="notif-title">
            <a :href="notif.url" target="_blank" rel="noopener noreferrer">{{ notif.title }}</a>
          </div>
          <div class="notif-meta">
            <span class="badge-source">{{ notif.source }}</span>
            <span v-if="notif.price !== null">${{ notif.price }}</span>
          </div>
        </div>
      </template>

      <p v-if="!lastScanResult.notifications.length && !lastScanResult.errors.length" class="no-results">
        No notifications generated
      </p>
    </div>
  </div>
</template>

<style scoped>
.tester-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tester-header {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.tester-header h3 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
}

.hint {
  font-size: 0.75rem;
  color: rgba(219, 227, 240, 0.45);
  font-style: italic;
}

.badge {
  font-size: 0.7rem;
  padding: 2px 8px;
  border-radius: 999px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.badge-enabled {
  background: rgba(130, 255, 180, 0.15);
  color: #82ffb4;
  border: 1px solid rgba(130, 255, 180, 0.3);
}

.badge-disabled {
  background: rgba(255, 160, 100, 0.1);
  color: #ffa064;
  border: 1px solid rgba(255, 160, 100, 0.3);
}

.field-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 8px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 0.82rem;
}

.field-wide {
  grid-column: span 2;
}

.field span {
  color: rgba(219, 227, 240, 0.6);
  font-size: 0.75rem;
}

.field input,
.field select {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(154, 173, 201, 0.2);
  border-radius: 6px;
  padding: 6px 8px;
  color: inherit;
  font: inherit;
  font-size: 0.82rem;
}

.checkbox-field {
  flex-direction: row;
  align-items: center;
  gap: 8px;
  padding-top: 18px;
}

.payload-preview {
  font-size: 0.78rem;
}

.payload-preview summary {
  cursor: pointer;
  color: rgba(219, 227, 240, 0.5);
  user-select: none;
}

.payload-preview pre {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(154, 173, 201, 0.15);
  border-radius: 6px;
  padding: 10px;
  overflow-x: auto;
  font-size: 0.75rem;
  line-height: 1.5;
  margin: 6px 0 0;
}

.tester-actions {
  display: flex;
  gap: 8px;
}

.run-btn {
  padding: 8px 20px;
  font-size: 0.85rem;
  border-radius: 8px;
}

.msg {
  font-size: 0.82rem;
  margin: 0;
  padding: 6px 10px;
  border-radius: 6px;
}

.msg-error {
  background: rgba(255, 80, 80, 0.1);
  color: #ff8080;
  border: 1px solid rgba(255, 80, 80, 0.2);
}

.msg-success {
  background: rgba(80, 255, 150, 0.1);
  color: #50ff96;
  border: 1px solid rgba(80, 255, 150, 0.2);
}

.result-box {
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(154, 173, 201, 0.15);
  border-radius: 8px;
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 0.8rem;
}

.result-label {
  margin: 0 0 4px;
  font-weight: 600;
  font-size: 0.78rem;
  color: rgba(219, 227, 240, 0.5);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.section-label {
  margin: 8px 0 2px;
  font-size: 0.75rem;
  font-weight: 600;
  color: rgba(219, 227, 240, 0.45);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.result-row {
  display: flex;
  gap: 10px;
}

.result-key {
  color: rgba(219, 227, 240, 0.5);
  min-width: 80px;
}

.outcome-failed { color: #ff8080; }

.notif-card {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(154, 173, 201, 0.1);
  border-radius: 6px;
  padding: 8px 10px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.notif-title a {
  color: #7dfdd4;
  text-decoration: none;
  font-size: 0.82rem;
}

.notif-title a:hover {
  text-decoration: underline;
}

.notif-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.75rem;
  color: rgba(219, 227, 240, 0.5);
}

.badge-source {
  font-size: 0.7rem;
  padding: 1px 6px;
  border-radius: 4px;
  background: rgba(62, 167, 255, 0.15);
  color: #3ea7ff;
  border: 1px solid rgba(62, 167, 255, 0.25);
}

.no-results {
  margin: 0;
  color: rgba(219, 227, 240, 0.4);
  font-style: italic;
}
</style>
