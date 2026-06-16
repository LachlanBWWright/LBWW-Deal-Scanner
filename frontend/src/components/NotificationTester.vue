<script setup lang="ts">
import { computed, ref } from "vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  useTestingConsole,
  type DeliveryMode,
  type TestingNotificationRequestBody,
} from "../composables/useTestingConsole";
import { formatDisplayLabel } from "../utils/displayLabels";

const {
  capabilities,
  loading: testingLoading,
  lastNotificationResult,
  sending,
  errorMessage,
  successMessage,
  sendTestNotification,
} = useTestingConsole();

type NotificationKind = "deal" | "error";

const kind = ref<NotificationKind>("deal");
const source = ref("ebay");
const title = ref("");
const url = ref("");
const price = ref<number | "">("");
const imageUrl = ref("");
const queryType = ref("ebay");
const queryId = ref("");
const dmOnly = ref(false);
const message = ref("");
const deliveryMode = ref<DeliveryMode>("normal");
const targetDiscordUserId = ref("");

const dealSources = [
  "cashConverters",
  "ebay",
  "gumtree",
  "salvos",
  "steamMarket",
  "csTrade",
  "lootFarm",
  "tradeIt",
];

const queryTypes = [
  "cashConverters",
  "ebay",
  "gumtree",
  "salvos",
  "steamMarket",
  "csTradeBot",
  "csMarket",
];

const deliveryModes: { value: DeliveryMode; label: string }[] = [
  { value: "normal", label: "Normal routing" },
  { value: "guildChannelOnly", label: "Guild channel only" },
  { value: "subscribedDMs", label: "Subscribed DMs" },
  { value: "specificUserDM", label: "Specific Discord user DM" },
];

type AllowedQueryType =
  | "cashConverters"
  | "ebay"
  | "gumtree"
  | "salvos"
  | "csMarket"
  | "steamMarket"
  | "csTradeBot";

const allowedQueryTypes: AllowedQueryType[] = [
  "cashConverters",
  "ebay",
  "gumtree",
  "salvos",
  "csMarket",
  "steamMarket",
  "csTradeBot",
];

function isAllowedQueryType(value: string): value is AllowedQueryType {
  return allowedQueryTypes.some((type) => type === value);
}

const payloadPreview = computed(() => {
  if (kind.value === "error") {
    return JSON.stringify(
      {
        kind: "error",
        source: source.value || "TestScanner",
        message: message.value || "(empty)",
        deliveryMode: deliveryMode.value,
      },
      null,
      2,
    );
  }

  const payload: Record<string, unknown> = {
    kind: "deal",
    source: source.value,
    title: title.value || "(empty)",
    url: url.value || "(empty)",
    deliveryMode: deliveryMode.value,
  };
  if (price.value !== "") payload.price = price.value;
  if (imageUrl.value) payload.imageUrl = imageUrl.value;
  if (queryId.value) {
    payload.query = { type: queryType.value, id: queryId.value, dmOnly: dmOnly.value };
  }
  if (deliveryMode.value === "specificUserDM" && targetDiscordUserId.value) {
    payload.targetDiscordUserId = targetDiscordUserId.value;
  }
  return JSON.stringify(payload, null, 2);
});

const testingEnabled = computed(() => capabilities.value?.testingEnabled === true);
const testingStatusLabel = computed(() => {
  if (testingLoading.value) return "Testing loading...";
  return testingEnabled.value ? "Testing enabled" : "Testing disabled";
});
const testingStatusVariant = computed(() => {
  if (testingLoading.value) return "secondary";
  return testingEnabled.value ? "success" : "warning";
});

function updateKind(value: string) {
  if (value === "deal" || value === "error") {
    kind.value = value;
  }
}

function updateDeliveryMode(value: string) {
  const matched = deliveryModes.find((mode) => mode.value === value);
  if (matched) {
    deliveryMode.value = matched.value;
  }
}

function onPriceInput(value: string) {
  if (!value.trim()) {
    price.value = "";
    return;
  }
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) return;
  price.value = parsed;
}

async function send() {
  if (!testingEnabled.value) return;

  if (kind.value === "error") {
    const errorBody: TestingNotificationRequestBody = {
      kind: "error",
      source: source.value || "TestScanner",
      message: message.value || "Test error",
      deliveryMode: deliveryMode.value,
    };

    await sendTestNotification(errorBody);
    return;
  }

  const body: TestingNotificationRequestBody = {
    kind: "deal",
    source: source.value,
    title: title.value,
    url: url.value,
    deliveryMode: deliveryMode.value,
  };
  if (price.value !== "") body.price = Number(price.value);
  if (imageUrl.value) body.imageUrl = imageUrl.value;
  if (queryId.value && isAllowedQueryType(queryType.value)) {
    body.query = {
      type: queryType.value,
      id: queryId.value,
      dmOnly: dmOnly.value,
    };
  }
  if (deliveryMode.value === "specificUserDM" && targetDiscordUserId.value) {
    body.targetDiscordUserId = targetDiscordUserId.value;
  }

  await sendTestNotification(body);
}
</script>
<script lang="ts">
export default {};
</script>

<template>
  <Card class="flex min-h-0 flex-1 flex-col">
    <CardContent class="flex min-h-0 flex-1 flex-col space-y-5 pt-6">
      <div class="flex flex-wrap items-center gap-2">
        <h2 class="text-lg font-semibold tracking-tight">Notification Tester</h2>
        <Badge :variant="testingStatusVariant">{{ testingStatusLabel }}</Badge>
      </div>
      <p class="text-sm text-muted-foreground">Build deal and error payloads using production-like controls.</p>

      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <div class="space-y-1.5">
          <Label>Kind</Label>
          <Select :model-value="kind" @update:model-value="updateKind">
            <option value="deal">Deal</option>
            <option value="error">Error</option>
          </Select>
        </div>

        <div class="space-y-2 md:col-span-2 lg:col-span-3">
          <Label>Source</Label>
          <Tabs v-model="source">
            <TabsList class="flex h-auto w-full flex-wrap gap-1 bg-muted/60 p-1">
              <TabsTrigger
                v-for="s in dealSources"
                :key="s"
                :value="s"
                class="min-h-9 min-w-0 flex-1 rounded-md px-3 py-2 text-xs sm:text-sm data-[state=active]:bg-primary data-[state=active]:text-primary-foreground data-[state=active]:shadow-sm"
              >
                {{ formatDisplayLabel(s) }}
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>

        <template v-if="kind === 'deal'">
          <div class="space-y-1.5 md:col-span-2">
            <Label>Title</Label>
            <Input v-model="title" type="text" placeholder="Deal title" />
          </div>

          <div class="space-y-1.5 md:col-span-2 lg:col-span-3">
            <Label>URL</Label>
            <Input v-model="url" type="url" placeholder="https://..." />
          </div>

          <div class="space-y-1.5">
            <Label>Price</Label>
            <Input :model-value="String(price)" type="number" step="0.01" min="0" placeholder="0.00" @update:model-value="onPriceInput" />
          </div>

          <div class="space-y-1.5 md:col-span-2">
            <Label>Image URL</Label>
            <Input v-model="imageUrl" type="url" placeholder="https://..." />
          </div>

          <div class="space-y-1.5">
            <Label>Query type</Label>
            <Select v-model="queryType">
              <option v-for="qt in queryTypes" :key="qt" :value="qt">{{ formatDisplayLabel(qt) }}</option>
            </Select>
          </div>

          <div class="space-y-1.5">
            <Label>Query ID</Label>
            <Input v-model="queryId" type="text" placeholder="(optional)" />
          </div>

          <div class="flex items-end">
            <div class="flex h-10 items-center gap-2 rounded-md border border-border/70 bg-muted/30 px-3 py-2">
              <Checkbox v-model="dmOnly" />
              <Label>DM only</Label>
            </div>
          </div>
        </template>

        <template v-else>
          <div class="space-y-1.5 md:col-span-2 lg:col-span-3">
            <Label>Error message</Label>
            <Input v-model="message" type="text" placeholder="Error description" />
          </div>
        </template>

        <div class="space-y-1.5">
          <Label>Delivery mode</Label>
          <Select :model-value="deliveryMode" @update:model-value="updateDeliveryMode">
            <option v-for="mode in deliveryModes" :key="mode.value" :value="mode.value">
              {{ mode.label }}
            </option>
          </Select>
        </div>

        <div v-if="deliveryMode === 'specificUserDM'" class="space-y-1.5">
          <Label>Discord user ID</Label>
          <Input v-model="targetDiscordUserId" type="text" placeholder="123456789012345678" />
        </div>
      </div>

      <details class="rounded-md border border-border/70 bg-background/60 p-3 text-sm">
        <summary class="cursor-pointer font-medium">Payload preview</summary>
        <pre class="mt-2 overflow-x-auto rounded-md bg-slate-950 p-3 text-xs text-slate-100">{{ payloadPreview }}</pre>
      </details>

      <div class="flex flex-wrap gap-2">
        <Button :disabled="sending || !testingEnabled" @click="send">
          {{ sending ? "Sending..." : "Send notification" }}
        </Button>
      </div>

      <p v-if="errorMessage" class="rounded-md border border-red-500/40 bg-red-500/10 px-3 py-2 text-sm font-medium text-red-200">{{ errorMessage }}</p>
      <p v-if="successMessage" class="rounded-md border border-emerald-500/40 bg-emerald-500/10 px-3 py-2 text-sm font-medium text-emerald-200">{{ successMessage }}</p>

      <div v-if="lastNotificationResult" class="space-y-3 rounded-xl border border-border/70 bg-card p-4">
        <p class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Last result</p>
        <div class="grid gap-2 text-sm sm:grid-cols-2">
          <p><strong>Duration:</strong> {{ lastNotificationResult.durationMs }}ms</p>
          <p><strong>Error:</strong> {{ lastNotificationResult.error ?? "none" }}</p>
        </div>
        <div v-for="outcome in lastNotificationResult.outcomes" :key="outcome.provider" class="flex flex-wrap items-center gap-2 text-sm">
          <strong>{{ outcome.provider }}</strong>
          <Badge
            :variant="
              outcome.status === 'sent'
                ? 'success'
                : outcome.status === 'failed'
                ? 'destructive'
                : 'warning'
            "
          >
            {{ outcome.status }}{{ outcome.reason ? ` - ${outcome.reason}` : "" }}
          </Badge>
        </div>
      </div>
    </CardContent>
  </Card>
</template>
