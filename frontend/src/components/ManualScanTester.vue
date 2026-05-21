<script setup lang="ts">
import { computed, ref } from "vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { useTestingConsole } from "../composables/useTestingConsole";
import type { TestingScanRequestBody } from "../composables/useTestingConsole";
import { formatDisplayLabel } from "../utils/displayLabels";

type ScannerType = TestingScanRequestBody["type"];

const defaultScannerTypes: ScannerType[] = [
  "cashConverters",
  "ebay",
  "gumtree",
  "salvos",
  "csMarket",
  "steamMarket",
  "csTradeBot",
];

const {
  capabilities,
  loading: testingLoading,
  lastScanResult,
  scanning,
  errorMessage,
  successMessage,
  runTestScan,
} = useTestingConsole();

const scannerType = ref<ScannerType>("ebay");
const notifyToggle = ref(false);
const validationMessage = ref<string | null>(null);

const url = ref("");
const maxPrice = ref("");
const minPrice = ref("");
const minFloat = ref("");
const maxFloat = ref("");
const displayUrl = ref("");
const name = ref("");
const requiredPhrases = ref("");
const excludePhrases = ref("");

const testingEnabled = computed(() => capabilities.value?.testingEnabled === true);
const testingStatusLabel = computed(() => {
  if (testingLoading.value) return "Testing loading...";
  return testingEnabled.value ? "Testing enabled" : "Testing disabled";
});
const testingStatusVariant = computed(() => {
  if (testingLoading.value) return "secondary";
  return testingEnabled.value ? "success" : "warning";
});
const supportedScannerTypes = computed<ScannerType[]>(() => {
  const types = capabilities.value?.availableScannerTypes ?? [];
  const scannerTypes = types.filter(isScannerType);
  return scannerTypes.length > 0 ? scannerTypes : defaultScannerTypes;
});

const scannerLabel = computed(() => formatDisplayLabel(scannerType.value));
const payloadPreview = computed(() => JSON.stringify(buildPayload(), null, 2));

const dealScanPreview = computed(() => ({
  type: scannerType.value,
  identity: identityPreview(),
  filters: filterPreview(),
  deliveryMode: notifyToggle.value ? "publish notifications" : "preview only",
  generatedPayload: buildPayload(),
}));

function isScannerType(value: string): value is ScannerType {
  return defaultScannerTypes.some((scannerTypeValue) => scannerTypeValue === value);
}

function optionalNumber(value: unknown) {
  const trimmed = String(value ?? "").trim();
  if (!trimmed) return undefined;
  return Number(trimmed);
}

function requiredNumber(value: unknown, label: string) {
  const trimmed = String(value ?? "").trim();
  if (!trimmed) {
    validationMessage.value = `${label} is required.`;
    return null;
  }
  const parsed = Number(trimmed);
  if (!Number.isFinite(parsed)) {
    validationMessage.value = `${label} must be a number.`;
    return null;
  }
  return parsed;
}

function requiredText(value: string, label: string) {
  const trimmed = value.trim();
  if (!trimmed) {
    validationMessage.value = `${label} is required.`;
    return null;
  }
  return trimmed;
}

function buildPayload(): Record<string, unknown> {
  switch (scannerType.value) {
    case "ebay":
    case "gumtree":
      return {
        url: url.value,
        ...(optionalNumber(maxPrice.value) !== undefined && {
          maxPrice: optionalNumber(maxPrice.value),
        }),
      };
    case "cashConverters":
      return {
        url: url.value,
        ...(requiredPhrases.value.trim() && { requiredPhrases: requiredPhrases.value }),
        ...(excludePhrases.value.trim() && { excludePhrases: excludePhrases.value }),
      };
    case "salvos":
      return {
        name: name.value,
        ...(optionalNumber(minPrice.value) !== undefined && { minPrice: optionalNumber(minPrice.value) }),
        ...(optionalNumber(maxPrice.value) !== undefined && { maxPrice: optionalNumber(maxPrice.value) }),
      };
    case "csMarket":
      return {
        url: url.value,
        displayUrl: displayUrl.value || url.value,
        ...(optionalNumber(maxPrice.value) !== undefined && { maxPrice: optionalNumber(maxPrice.value) }),
        ...(optionalNumber(maxFloat.value) !== undefined && { maxFloat: optionalNumber(maxFloat.value) }),
      };
    case "steamMarket":
      return {
        name: name.value,
        displayUrl: displayUrl.value || name.value,
        ...(optionalNumber(maxPrice.value) !== undefined && { maxPrice: optionalNumber(maxPrice.value) }),
      };
    case "csTradeBot":
      return {
        name: name.value,
        ...(optionalNumber(minFloat.value) !== undefined && { minFloat: optionalNumber(minFloat.value) }),
        ...(optionalNumber(maxFloat.value) !== undefined && { maxFloat: optionalNumber(maxFloat.value) }),
        ...(optionalNumber(maxPrice.value) !== undefined && { maxPrice: optionalNumber(maxPrice.value) }),
      };
    default:
      return {};
  }
}

function identityPreview() {
  if (scannerType.value === "salvos" || scannerType.value === "steamMarket" || scannerType.value === "csTradeBot") {
    return name.value || "(missing name)";
  }
  return url.value || "(missing URL)";
}

function filterPreview() {
  return {
    maxPrice: optionalNumber(maxPrice.value),
    minPrice: optionalNumber(minPrice.value),
    minFloat: optionalNumber(minFloat.value),
    maxFloat: optionalNumber(maxFloat.value),
    requiredPhrases: requiredPhrases.value || undefined,
    excludePhrases: excludePhrases.value || undefined,
  };
}

function resetFields() {
  url.value = "";
  maxPrice.value = "";
  minPrice.value = "";
  minFloat.value = "";
  maxFloat.value = "";
  displayUrl.value = "";
  name.value = "";
  requiredPhrases.value = "";
  excludePhrases.value = "";
  validationMessage.value = null;
}

function onTypeChange(event: Event) {
  const target = event.target;
  if (!(target instanceof HTMLSelectElement)) return;
  if (isScannerType(target.value)) {
    scannerType.value = target.value;
    resetFields();
  }
}

function validate() {
  validationMessage.value = null;
  if (scannerType.value === "ebay" || scannerType.value === "gumtree" || scannerType.value === "cashConverters") {
    requiredText(url.value, "Query URL");
    return !validationMessage.value;
  }
  if (scannerType.value === "salvos") {
    requiredText(name.value, "Item name");
    return !validationMessage.value;
  }
  if (scannerType.value === "csMarket") {
    requiredText(url.value, "Search URL");
    if (validationMessage.value) return false;
    requiredNumber(maxPrice.value, "Max price");
    if (validationMessage.value) return false;
    requiredNumber(maxFloat.value, "Max float");
    return !validationMessage.value;
  }
  if (scannerType.value === "steamMarket") {
    requiredText(name.value, "Query name");
    if (validationMessage.value) return false;
    requiredNumber(maxPrice.value, "Max price");
    return !validationMessage.value;
  }
  if (scannerType.value === "csTradeBot") {
    requiredText(name.value, "Item name");
    if (validationMessage.value) return false;
    requiredNumber(minFloat.value, "Min float");
    if (validationMessage.value) return false;
    requiredNumber(maxFloat.value, "Max float");
    if (validationMessage.value) return false;
    requiredNumber(maxPrice.value, "Max price");
    return !validationMessage.value;
  }
  return true;
}

async function run() {
  if (!validate()) return;

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
  <Card class="flex min-h-0 flex-1 flex-col">
    <CardContent class="flex min-h-0 flex-1 flex-col space-y-5 pt-6">
      <div class="flex flex-wrap items-center gap-2">
        <h2 class="text-lg font-semibold tracking-tight">Manual Scan Tester</h2>
        <Badge :variant="testingStatusVariant">{{ testingStatusLabel }}</Badge>
      </div>
      <p class="text-sm text-muted-foreground">Temporary scan input. No saved query will be modified.</p>

      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <div class="space-y-1.5">
          <Label>Scanner type</Label>
          <Select :model-value="scannerType" @change="onTypeChange">
            <option v-for="t in supportedScannerTypes" :key="t" :value="t">{{ formatDisplayLabel(t) }}</option>
          </Select>
        </div>

        <div class="flex items-end md:col-span-1">
          <div class="flex h-10 items-center gap-2 rounded-md border border-border/70 bg-muted/30 px-3 py-2">
            <Checkbox v-model="notifyToggle" />
            <Label>Publish notifications</Label>
          </div>
        </div>

        <template v-if="scannerType === 'ebay' || scannerType === 'gumtree'">
          <div class="space-y-1.5 md:col-span-2">
            <Label>Query URL</Label>
            <Input v-model="url" type="url" placeholder="https://..." />
          </div>
          <div class="space-y-1.5">
            <Label>Max price</Label>
            <Input v-model="maxPrice" type="number" step="0.01" min="0" />
          </div>
        </template>

        <template v-if="scannerType === 'cashConverters'">
          <div class="space-y-1.5 md:col-span-2">
            <Label>Query URL</Label>
            <Input v-model="url" type="url" placeholder="https://..." />
          </div>
          <div class="space-y-1.5">
            <Label>Required phrases</Label>
            <Input v-model="requiredPhrases" type="text" placeholder="(optional)" />
          </div>
          <div class="space-y-1.5">
            <Label>Exclude phrases</Label>
            <Input v-model="excludePhrases" type="text" placeholder="(optional)" />
          </div>
        </template>

        <template v-if="scannerType === 'salvos' || scannerType === 'steamMarket' || scannerType === 'csTradeBot'">
          <div class="space-y-1.5">
            <Label>{{ scannerType === "steamMarket" ? "Query name" : "Item name" }}</Label>
            <Input v-model="name" type="text" placeholder="e.g. iPhone" />
          </div>
        </template>

        <template v-if="scannerType === 'csMarket' || scannerType === 'steamMarket'">
          <div v-if="scannerType === 'csMarket'" class="space-y-1.5 md:col-span-2">
            <Label>Search URL</Label>
            <Input v-model="url" type="url" placeholder="https://..." />
          </div>
          <div class="space-y-1.5 md:col-span-2">
            <Label>Display URL</Label>
            <Input v-model="displayUrl" type="text" placeholder="Shown in notification" />
          </div>
        </template>

        <template v-if="scannerType === 'salvos'">
          <div class="space-y-1.5">
            <Label>Min price</Label>
            <Input v-model="minPrice" type="number" step="0.01" min="0" />
          </div>
          <div class="space-y-1.5">
            <Label>Max price</Label>
            <Input v-model="maxPrice" type="number" step="0.01" min="0" />
          </div>
        </template>

        <template v-if="scannerType === 'csMarket' || scannerType === 'steamMarket' || scannerType === 'csTradeBot'">
          <div v-if="scannerType === 'csTradeBot'" class="space-y-1.5">
            <Label>Min float</Label>
            <Input v-model="minFloat" type="number" step="0.0001" min="0" max="1" />
          </div>
          <div v-if="scannerType !== 'steamMarket'" class="space-y-1.5">
            <Label>Max float</Label>
            <Input v-model="maxFloat" type="number" step="0.0001" min="0" max="1" />
          </div>
          <div class="space-y-1.5">
            <Label>Max price</Label>
            <Input v-model="maxPrice" type="number" step="0.01" min="0" />
          </div>
        </template>
      </div>

      <p v-if="validationMessage" class="rounded-md border border-red-500/40 bg-red-500/10 px-3 py-2 text-sm font-medium text-red-200">{{ validationMessage }}</p>

      <div class="grid gap-2 rounded-lg border border-border/70 bg-muted/30 p-4 text-sm sm:grid-cols-2">
        <p><strong>Type:</strong> {{ scannerLabel }}</p>
        <p><strong>Identity:</strong> {{ dealScanPreview.identity }}</p>
        <p><strong>Delivery:</strong> {{ dealScanPreview.deliveryMode }}</p>
        <p class="sm:col-span-2 break-all"><strong>Filters:</strong> {{ JSON.stringify(dealScanPreview.filters) }}</p>
      </div>

      <details class="rounded-md border border-border/70 bg-background/60 p-3 text-sm">
        <summary class="cursor-pointer font-medium">Generated payload</summary>
        <pre class="mt-2 overflow-x-auto rounded-md bg-slate-950 p-3 text-xs text-slate-100">{{ payloadPreview }}</pre>
      </details>

      <div class="flex flex-wrap gap-2">
        <Button :disabled="scanning" @click="run">{{ scanning ? "Running scan..." : "Run scan" }}</Button>
      </div>

      <p v-if="errorMessage" class="rounded-md border border-red-500/40 bg-red-500/10 px-3 py-2 text-sm font-medium text-red-200">{{ errorMessage }}</p>
      <p v-if="successMessage" class="rounded-md border border-emerald-500/40 bg-emerald-500/10 px-3 py-2 text-sm font-medium text-emerald-200">{{ successMessage }}</p>

      <div v-if="lastScanResult" class="space-y-3 rounded-xl border border-border/70 bg-card p-4">
        <p class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Scan result</p>

        <div class="grid gap-2 text-sm sm:grid-cols-3">
          <p><strong>Duration:</strong> {{ lastScanResult.durationMs }}ms</p>
          <p><strong>Published:</strong> {{ lastScanResult.notificationsPublished ? "Yes" : "No" }}</p>
          <p><strong>Items found:</strong> {{ lastScanResult.items.length }}</p>
        </div>

        <template v-if="lastScanResult.errors.length">
          <p class="text-sm font-semibold text-red-200">Errors</p>
          <p v-for="(err, idx) in lastScanResult.errors" :key="idx" class="text-sm text-red-200">{{ err }}</p>
        </template>

        <template v-if="lastScanResult.items.length">
          <p class="text-sm font-semibold">Items found ({{ lastScanResult.items.length }})</p>
          <div class="space-y-2">
            <div
              v-for="item in lastScanResult.items"
              :key="item.url + item.title"
              class="rounded-lg border p-3"
              :class="item.passedFilters ? 'border-emerald-500/40 bg-emerald-500/10' : 'border-amber-500/40 bg-amber-500/10'"
            >
              <a :href="item.url" target="_blank" rel="noopener noreferrer" class="text-sm font-semibold text-blue-300 hover:underline">{{ item.title }}</a>
              <div class="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                <Badge variant="secondary">{{ formatDisplayLabel(item.source) }}</Badge>
                <span v-if="item.price !== null">${{ item.price }}</span>
                <Badge :variant="item.passedFilters ? 'success' : 'warning'">{{ item.passedFilters ? "passes filters" : "filtered out" }}</Badge>
              </div>
              <p v-if="item.filterReason" class="mt-1 text-xs text-muted-foreground">{{ item.filterReason }}</p>
            </div>
          </div>
        </template>

        <template v-if="lastScanResult.notifications.length">
          <p class="text-sm font-semibold">Filter-passing notifications ({{ lastScanResult.notifications.length }})</p>
          <div class="space-y-2">
            <div v-for="notif in lastScanResult.notifications" :key="notif.url" class="rounded-lg border border-border/70 bg-muted/20 p-3">
              <a :href="notif.url" target="_blank" rel="noopener noreferrer" class="text-sm font-semibold text-blue-300 hover:underline">{{ notif.title }}</a>
              <div class="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                <Badge variant="secondary">{{ formatDisplayLabel(notif.source) }}</Badge>
                <span v-if="notif.price !== null">${{ notif.price }}</span>
              </div>
            </div>
          </div>
        </template>

        <p
          v-if="!lastScanResult.items.length && !lastScanResult.notifications.length && !lastScanResult.errors.length"
          class="text-sm italic text-muted-foreground"
        >
          No items found
        </p>
      </div>
    </CardContent>
  </Card>
</template>
