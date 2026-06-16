# Backend Notification Refactor Plan

## Goal

Refactor `server/` so deal scanning is independent from Discord delivery. Scanners should discover deal events and hand them to a notification boundary. Discord should become one notification adapter, with room to add desktop notifications, email, webhooks, or other delivery channels without changing scanner logic.

## Current Coupling

The current server mixes scanning and Discord concerns in a few places:

- `server/scanners/siteScanners/*` imports `sendToChannel`, reads Discord channel IDs and role IDs from `globals`, and formats Discord mentions inline.
- `server/functions/sendToChannel.ts` handles Discord channel posts, DMs, action buttons, query subscription behavior, and query lookup.
- `server/functions/handleError.ts` reports errors through `sendToChannel`.
- `server/discordBot.ts` starts the background scanner loop only after the Discord client is ready.
- Feature flags and routing configuration are Discord-shaped, for example `EBAY_CHANNEL_ID`, `EBAY_ROLE_ID`, `ERROR_CHANNEL_ID`.

This makes scanners hard to reuse in non-Discord contexts and means a missing Discord config disables both Discord and background scanning.

## Target Shape

Create four clear layers:

- `scanning`: browser/page orchestration, site-specific extraction, query selection, new-item detection, and scan scheduling.
- `deals`: shared domain types for found deals, source names, query references, severity/event types, and message content.
- `notifications`: provider-neutral interfaces, notification routing, subscription lookup, fan-out, formatting hooks, and error handling.
- `integrations/discord`: Discord client startup, commands, interaction handling, Discord-specific formatting, channel/DM sending, and action buttons.

Suggested directory structure:

```text
server/
  scanning/
    scannerRuntime.ts
    scanRegistry.ts
    siteScanners/
  deals/
    types.ts
    dedupe.ts
    messages.ts
  notifications/
    types.ts
    notificationService.ts
    routes.ts
    subscribers.ts
    errorReporter.ts
  integrations/
    discord/
      discordBot.ts
      discordNotifier.ts
      discordClient.ts
      commandManager/
      actionRegistry.ts
  index.ts
```

Names can be adjusted during implementation, but the important rule is one-way dependency flow:

```text
index -> scanning
index -> notifications
notifications -> integrations/*
scanning -> deals
scanning -> notifications/types only
integrations/discord -> command/query management
```

`scanning` must not import `discord.js`, `DiscordJSClient`, `sendToChannel`, Discord channel IDs, or Discord role IDs.

## Core Interfaces

Add a provider-neutral notification contract:

```ts
export type NotificationTarget =
  | { type: "source"; source: DealSource }
  | { type: "query"; queryType: QueryType; queryId: string }
  | { type: "errors" };

export interface DealNotification {
  kind: "deal";
  source: DealSource;
  title: string;
  url: string;
  price?: number;
  imageUrl?: string;
  query?: {
    type: QueryType;
    id: string;
    dmOnly?: boolean;
  };
  tags?: string[];
}

export interface ErrorNotification {
  kind: "error";
  source: string;
  message: string;
  stack?: string;
}

export type AppNotification = DealNotification | ErrorNotification;

export interface NotificationProvider {
  name: string;
  isEnabled(): boolean;
  send(notification: AppNotification): Promise<void>;
}
```

Add a service that owns fan-out:

```ts
export interface NotificationService {
  publish(notification: AppNotification): Promise<void>;
}
```

Scanners should call `notificationService.publish(...)` or return `DealNotification[]` to the runtime, which then publishes them. Prefer returning notifications from scanners if the scanner API can be changed cleanly, because it makes scanner tests simpler and keeps side effects in the runtime.

## Refactor Steps

### 1. Introduce Domain Types

Create `server/deals/types.ts` and move shared concepts there:

- `DealSource`: `cashConverters`, `ebay`, `gumtree`, `salvos`, `steamMarket`, `csTrade`, `lootFarm`, `tradeIt`.
- `QueryType`: existing query type strings used by `sendToChannel`.
- `DealNotification`, `ErrorNotification`, and `AppNotification`.
- Optional helpers for source labels and default notification text.

Keep the initial shape close to existing message data: title/name, formatted price or numeric price, URL, image URL, query metadata, source.

### 2. Add Notification Service

Create `server/notifications/notificationService.ts`:

- Accepts an array of `NotificationProvider`.
- Filters disabled providers.
- Sends to each enabled provider.
- Catches provider-specific failures so one provider does not block others.
- Logs failures and exposes testable behavior.

Create `server/notifications/errorReporter.ts` to replace direct `handleError -> sendToChannel` usage with provider-neutral `ErrorNotification` publishing.

### 3. Move Discord Sending Behind a Provider

Move `server/functions/sendToChannel.ts` behavior into `server/integrations/discord/discordNotifier.ts`.

The Discord provider should own:

- Channel routing from source/error notification to Discord channel ID.
- Role mention formatting.
- Discord message text formatting.
- Attachments from `imageUrl`.
- DM subscription sends.
- `dmOnly` handling.
- Delete/subscribe/unsubscribe action button creation.
- `actionRegistry` cleanup.

Keep `getQueryByTypeAndId` either inside the Discord provider initially or move it to `notifications/subscribers.ts` if desktop notifications will share query subscription logic. For the first pass, keep it in Discord unless another provider needs it.

### 4. Decouple Scanner Startup From Discord Startup

Change `server/index.ts` so the API and scanning runtime are application-level services, not Discord-owned services.

Target behavior:

- API starts independently.
- Notification providers are constructed.
- Scanner runtime starts if scanning is enabled, regardless of whether Discord is configured.
- Discord bot starts if Discord config exists.
- Discord command handling remains available when Discord is enabled.

`server/discordBot.ts` should no longer import or call `startBackgroundScanLoop`.

### 5. Update Scanner APIs

Change scanner functions from "scan and send Discord message" to either:

```ts
export async function scanEbay(page: Page): Promise<DealNotification[]>
```

or:

```ts
export async function scanEbay(
  page: Page,
  notifications: NotificationService,
): Promise<void>
```

Prefer the return-value approach where practical:

- Existing extraction tests can assert returned notifications.
- Runtime owns side effects.
- Desktop notification support becomes a runtime/provider concern.

Each scanner should still:

- Read source enablement flags until config is refactored.
- Load the next query.
- Extract the found item.
- Apply price/query constraints.
- Call `checkIfNew`.

Each scanner should stop:

- Importing `sendToChannel`.
- Reading Discord channel IDs.
- Reading Discord role IDs.
- Formatting Discord mentions.

### 6. Refactor Runtime Publishing

Update `server/scannerRuntime.ts`:

- Accept a `NotificationService` when starting or running scans.
- `runScanPass` collects each scanner's returned notifications.
- Publish each notification through the service.
- Keep scanner failure isolation through `handleScan`, but report errors through `ErrorNotification`.

Example shape:

```ts
type ScannerTask = {
  name: string;
  scan: () => Promise<AppNotification[]>;
};
```

This also gives a natural place for a scanner registry instead of hard-coded scan order directly in `runScanPass`.

### 7. Normalize Configuration

Keep current env/database fields in the first implementation to reduce migration risk, but hide Discord-specific fields behind the Discord provider.

Then add a second pass to introduce provider-neutral config:

- Source enablement: `sources.cashConverters.enabled`, `sources.ebay.enabled`, etc.
- Notification routing: source -> provider target, for example Discord channel ID or desktop channel.
- Provider enablement: `notifications.discord.enabled`, `notifications.desktop.enabled`.
- Error routing: `notifications.errors.targets`.

Do not rename Prisma fields in the first refactor unless required. That would combine architecture changes with a data migration.

### 8. Add Desktop Notification Adapter

After Discord is isolated, add a small proof-of-extension provider:

```text
server/integrations/desktop/desktopNotifier.ts
```

Implementation options:

- Local development: use a lightweight library or OS command wrapper.
- Production/server deployments: no-op unless explicitly enabled.

This provider should consume the same `AppNotification` objects and prove that scanners need no changes.

### 9. Tests

Add or update tests around the boundaries:

- Scanner tests assert returned `DealNotification[]` and no Discord mocks are needed.
- `notificationService` tests cover fan-out, disabled providers, and provider failure isolation.
- Discord provider tests cover message formatting, channel routing, `dmOnly`, DM subscriptions, and action buttons.
- `scannerRuntime` tests cover scan pass publishing and error notification publishing.
- Existing command manager tests remain Discord-specific and move with the Discord integration if files are relocated.

Important regression checks:

- A missing Discord token does not prevent manual scans or background scans if scanning is enabled.
- `dmOnly` queries still suppress channel posts for Discord.
- Query delete and DM subscribe buttons still work.
- Error notifications still reach the configured Discord error channel when Discord is enabled.

## Suggested Implementation Order

1. Add domain notification types and `NotificationService`.
2. Wrap current `sendToChannel` behavior in a `DiscordNotificationProvider` without changing scanners yet.
3. Change `handleError` to publish `ErrorNotification`.
4. Update one low-risk scanner, such as eBay or Cash Converters, to return `DealNotification[]`.
5. Update `scannerRuntime` to publish returned notifications.
6. Convert remaining scanners in small batches.
7. Move Discord files into `integrations/discord`.
8. Decouple `startBackgroundScanLoop` from `startDiscordBot`.
9. Add desktop/no-op provider as an extension proof.
10. Clean up old `sendToChannel` imports and remove compatibility wrappers.

## Migration Strategy

Use an adapter-first migration to avoid a large rewrite:

- Keep `sendToChannel` temporarily as a thin compatibility wrapper around the Discord provider.
- Convert scanners one at a time.
- Remove the wrapper only after `rg "sendToChannel" server` finds no scanner/runtime usage.
- Keep existing env/database config names until all behavior is covered by tests.

This keeps Discord behavior stable while the scanner-facing API changes.

## Done Criteria

The refactor is complete when:

- No scanner imports `discord.js`, `DiscordJSClient`, `sendToChannel`, or Discord channel/role config.
- `server/discordBot.ts` does not start the scan loop.
- `scannerRuntime` depends only on scanner modules and provider-neutral notification interfaces.
- Discord notification behavior lives under `server/integrations/discord`.
- Adding a desktop notification provider requires registering a provider, not editing scanner files.
- Tests cover scanner output, notification fan-out, and Discord-specific delivery behavior.
