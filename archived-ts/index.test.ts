import { afterEach, describe, expect, it, vi } from "vitest";

const {
  startApiServer,
  startBackgroundScanLoop,
  runScheduledScanLoop,
  startDiscordBot,
  setNotificationService,
  initGlobals,
} = vi.hoisted(() => ({
  startApiServer: vi.fn(async () => ({ host: "127.0.0.1", port: 3001 })),
  startBackgroundScanLoop: vi.fn(),
  runScheduledScanLoop: vi.fn(async () => undefined),
  startDiscordBot: vi.fn(async () => undefined),
  setNotificationService: vi.fn(),
  initGlobals: vi.fn(async () => undefined),
}));

vi.mock("./api.js", () => ({
  startApiServer,
}));

vi.mock("./scannerRuntime.js", () => ({
  startBackgroundScanLoop,
  runScheduledScanLoop,
  setNotificationService,
}));

vi.mock("./integrations/discord/discordBot.js", () => ({
  startDiscordBot,
}));

vi.mock("./globals/Globals.js", () => ({
  initGlobals,
}));

vi.mock("./integrations/discord/discordNotifier.js", () => ({
  DiscordNotificationProvider: class {
    isEnabled() {
      return false;
    }
    send() {
      return Promise.resolve();
    }
  },
}));

vi.mock("./integrations/desktop/desktopNotifier.js", () => ({
  DesktopNotificationProvider: class {
    isEnabled() {
      return false;
    }
    send() {
      return Promise.resolve();
    }
  },
}));

vi.mock("./notifications/notificationService.js", () => ({
  DefaultNotificationService: function (...args: unknown[]) {
    void args;
    return {};
  },
}));

describe("index.ts", () => {
  const originalScheduledMode = process.env.SCHEDULED_SCANNER_MODE;
  const originalScheduledDuration = process.env.SCHEDULED_SCANNER_DURATION_MS;

  afterEach(() => {
    if (originalScheduledMode === undefined) {
      delete process.env.SCHEDULED_SCANNER_MODE;
    } else {
      process.env.SCHEDULED_SCANNER_MODE = originalScheduledMode;
    }

    if (originalScheduledDuration === undefined) {
      delete process.env.SCHEDULED_SCANNER_DURATION_MS;
    } else {
      process.env.SCHEDULED_SCANNER_DURATION_MS = originalScheduledDuration;
    }

    vi.clearAllMocks();
  });

  it("boots globals, notification service, API, scanner loop, and discord bot", async () => {
    vi.resetModules();
    await import("./index.js");

    expect(initGlobals).toHaveBeenCalledTimes(1);
    expect(setNotificationService).toHaveBeenCalledTimes(1);
    expect(startApiServer).toHaveBeenCalledTimes(1);
    expect(startBackgroundScanLoop).toHaveBeenCalledTimes(1);
    expect(startDiscordBot).toHaveBeenCalledTimes(1);
    expect(runScheduledScanLoop).not.toHaveBeenCalled();
  });

  it("runs only the scheduled scanner when scheduled mode is enabled", async () => {
    process.env.SCHEDULED_SCANNER_MODE = "true";
    process.env.SCHEDULED_SCANNER_DURATION_MS = "60000";

    vi.resetModules();
    await import("./index.js");

    expect(initGlobals).toHaveBeenCalledTimes(1);
    expect(setNotificationService).toHaveBeenCalledTimes(1);
    expect(runScheduledScanLoop).toHaveBeenCalledWith(60000);
    expect(startApiServer).not.toHaveBeenCalled();
    expect(startBackgroundScanLoop).not.toHaveBeenCalled();
    expect(startDiscordBot).not.toHaveBeenCalled();
  });
});
