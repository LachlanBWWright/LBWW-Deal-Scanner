import type { DealSource, QueryType } from "../deals/types.js";

// ─── Delivery mode ────────────────────────────────────────────────────────────

export type TestDeliveryMode =
  | "normal"
  | "guildChannelOnly"
  | "subscribedDMs"
  | "specificUserDM";

// ─── Notification request / result ───────────────────────────────────────────

export interface TestingNotificationRequest {
  kind: "deal" | "error";
  // Deal fields
  source?: DealSource;
  title?: string;
  url?: string;
  price?: number;
  imageUrl?: string;
  query?: {
    type: QueryType;
    id: string;
    dmOnly?: boolean;
  };
  tags?: string[];
  // Error fields
  message?: string;
  // Delivery control
  deliveryMode?: TestDeliveryMode;
  targetDiscordUserId?: string;
}

export interface ProviderOutcome {
  provider: string;
  status: "sent" | "skipped" | "disabled" | "failed";
  reason?: string;
}

export interface TestingNotificationResult {
  id: string;
  startedAt: string;
  finishedAt: string;
  durationMs: number;
  request: TestingNotificationRequest;
  outcomes: ProviderOutcome[];
  error: string | null;
}

// ─── Scan request / result ────────────────────────────────────────────────────

export interface ManualScannerInput {
  type: QueryType;
  payload: Record<string, unknown>;
  notify?: boolean;
}

export interface ScannerNotificationResult {
  source: string;
  title: string;
  url: string;
  price?: number;
  imageUrl?: string;
}

export interface ManualScanItemResult {
  source: string;
  title: string;
  url: string;
  price: number | null;
  imageUrl: string | null;
  passedFilters: boolean;
  filterReason: string | null;
}

export interface TestingScanResult {
  id: string;
  startedAt: string;
  finishedAt: string;
  durationMs: number;
  request: ManualScannerInput;
  items: ManualScanItemResult[];
  notifications: ScannerNotificationResult[];
  errors: string[];
  notificationsPublished: boolean;
}

// ─── Capabilities ─────────────────────────────────────────────────────────────

export interface TestingCapabilities {
  testingEnabled: boolean;
  notificationProviders: string[];
  discordConnected: boolean;
  availableQueryTypes: string[];
  availableScannerTypes: string[];
}
