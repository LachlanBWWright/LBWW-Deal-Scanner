<script setup lang="ts">
import { ref, computed } from "vue";
import { useTestingConsole, type DeliveryMode } from "../composables/useTestingConsole";

const {
  capabilities,
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

const presets = [
  {
    label: "eBay deal",
    apply() {
      kind.value = "deal";
      source.value = "ebay";
      title.value = "Test eBay item listing at $49.99";
      url.value = "https://www.ebay.com.au/itm/test";
      price.value = 49.99;
      imageUrl.value = "";
      message.value = "";
    },
  },
  {
    label: "Cash Converters deal",
    apply() {
      kind.value = "deal";
      source.value = "cashConverters";
      title.value = "Test Cash Converters item for $29.99";
      url.value = "https://www.cashconverters.com.au/products/test";
      price.value = 29.99;
      imageUrl.value = "";
      message.value = "";
    },
  },
  {
    label: "Steam Market deal",
    apply() {
      kind.value = "deal";
      source.value = "steamMarket";
      title.value = "Test Steam item listed at $12.50";
      url.value = "https://steamcommunity.com/market/listings/730/Test%20Item";
      price.value = 12.5;
      imageUrl.value = "";
      message.value = "";
    },
  },
  {
    label: "Gumtree deal",
    apply() {
      kind.value = "deal";
      source.value = "gumtree";
      title.value = "Test Gumtree listing for $75.00";
      url.value = "https://www.gumtree.com.au/s-ad/test/1234567890";
      price.value = 75.0;
      imageUrl.value = "";
      message.value = "";
    },
  },
  {
    label: "Scanner error",
    apply() {
      kind.value = "error";
      source.value = "TestScanner";
      message.value = "Simulated scanner error for testing";
      title.value = "";
      url.value = "";
      price.value = "";
    },
  },
];

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

const testingEnabled = computed(() => capabilities.value?.testingEnabled ?? false);

async function send() {
  if (!testingEnabled.value) return;

  if (kind.value === "error") {
    await sendTestNotification({
      kind: "error",
      source: source.value || "TestScanner",
      message: message.value || "Test error",
      deliveryMode: deliveryMode.value as "normal",
    });
    return;
  }

  const body: Parameters<typeof sendTestNotification>[0] = {
    kind: "deal",
    source: source.value as Parameters<typeof sendTestNotification>[0]["source"],
    title: title.value,
    url: url.value,
    deliveryMode: deliveryMode.value as "normal",
  };
  if (price.value !== "") body.price = Number(price.value);
  if (imageUrl.value) body.imageUrl = imageUrl.value;
  if (queryId.value) {
    body.query = {
      type: queryType.value as "ebay",
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

<template>
  <div class="tester-panel">
    <div class="tester-header">
      <h3>Notification Tester</h3>
      <div v-if="!testingEnabled" class="badge badge-disabled">Testing disabled</div>
      <div v-else class="badge badge-enabled">Testing enabled</div>
    </div>

    <div class="preset-row">
      <button
        v-for="preset in presets"
        :key="preset.label"
        class="preset-btn"
        @click="preset.apply()"
      >
        {{ preset.label }}
      </button>
    </div>

    <div class="field-grid">
      <label class="field">
        <span>Kind</span>
        <select v-model="kind">
          <option value="deal">Deal</option>
          <option value="error">Error</option>
        </select>
      </label>

      <label class="field">
        <span>Source</span>
        <select v-model="source">
          <option v-for="s in dealSources" :key="s" :value="s">{{ s }}</option>
        </select>
      </label>

      <template v-if="kind === 'deal'">
        <label class="field field-wide">
          <span>Title</span>
          <input v-model="title" type="text" placeholder="Deal title" />
        </label>

        <label class="field field-wide">
          <span>URL</span>
          <input v-model="url" type="url" placeholder="https://..." />
        </label>

        <label class="field">
          <span>Price</span>
          <input v-model.number="price" type="number" step="0.01" min="0" placeholder="0.00" />
        </label>

        <label class="field field-wide">
          <span>Image URL</span>
          <input v-model="imageUrl" type="url" placeholder="https://..." />
        </label>

        <label class="field">
          <span>Query type</span>
          <select v-model="queryType">
            <option v-for="qt in queryTypes" :key="qt" :value="qt">{{ qt }}</option>
          </select>
        </label>

        <label class="field">
          <span>Query ID</span>
          <input v-model="queryId" type="text" placeholder="(optional)" />
        </label>

        <label class="field checkbox-field">
          <input v-model="dmOnly" type="checkbox" />
          <span>DM only</span>
        </label>
      </template>

      <template v-else>
        <label class="field field-wide">
          <span>Error message</span>
          <input v-model="message" type="text" placeholder="Error description" />
        </label>
      </template>

      <label class="field">
        <span>Delivery mode</span>
        <select v-model="deliveryMode">
          <option
            v-for="mode in deliveryModes"
            :key="mode.value"
            :value="mode.value"
          >
            {{ mode.label }}
          </option>
        </select>
      </label>

      <label
        v-if="deliveryMode === 'specificUserDM'"
        class="field"
      >
        <span>Discord user ID</span>
        <input
          v-model="targetDiscordUserId"
          type="text"
          placeholder="123456789012345678"
        />
      </label>
    </div>

    <details class="payload-preview">
      <summary>Payload preview</summary>
      <pre>{{ payloadPreview }}</pre>
    </details>

    <div class="tester-actions">
      <button
        class="primary send-btn"
        :disabled="sending || !testingEnabled"
        @click="send"
      >
        {{ sending ? "Sending..." : "Send notification" }}
      </button>
    </div>

    <p v-if="errorMessage" class="msg msg-error">{{ errorMessage }}</p>
    <p v-if="successMessage" class="msg msg-success">{{ successMessage }}</p>

    <div v-if="lastNotificationResult" class="result-box">
      <p class="result-label">Last result</p>
      <div class="result-row">
        <span class="result-key">Duration</span>
        <span>{{ lastNotificationResult.durationMs }}ms</span>
      </div>
      <div class="result-row">
        <span class="result-key">Error</span>
        <span>{{ lastNotificationResult.error ?? "none" }}</span>
      </div>
      <div
        v-for="outcome in lastNotificationResult.outcomes"
        :key="outcome.provider"
        class="result-row"
      >
        <span class="result-key">{{ outcome.provider }}</span>
        <span
          :class="{
            'outcome-sent': outcome.status === 'sent',
            'outcome-failed': outcome.status === 'failed',
            'outcome-skipped': outcome.status === 'skipped' || outcome.status === 'disabled',
          }"
        >{{ outcome.status }}{{ outcome.reason ? ` — ${outcome.reason}` : "" }}</span>
      </div>
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
}

.tester-header h3 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
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

.preset-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.preset-btn {
  font-size: 0.78rem;
  padding: 4px 10px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(154, 173, 201, 0.2);
  color: inherit;
  cursor: pointer;
}

.preset-btn:hover {
  background: rgba(255, 255, 255, 0.1);
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

.send-btn {
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

.result-row {
  display: flex;
  gap: 10px;
}

.result-key {
  color: rgba(219, 227, 240, 0.5);
  min-width: 80px;
}

.outcome-sent { color: #82ffb4; }
.outcome-failed { color: #ff8080; }
.outcome-skipped { color: #ffa064; }
</style>
