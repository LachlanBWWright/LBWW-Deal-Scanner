import { describe, expect, it, vi } from "vitest";

const {
  startApiServer,
  startBackgroundScanLoop,
  startDiscordBot,
  setNotificationService,
  initGlobals,
} = vi.hoisted(() => ({
  startApiServer: vi.fn(async () => ({ host: "127.0.0.1", port: 3001 })),
  startBackgroundScanLoop: vi.fn(),
  startDiscordBot: vi.fn(async () => undefined),
  setNotificationService: vi.fn(),
  initGlobals: vi.fn(async () => undefined),
}));

vi.mock("./api.js", () => ({
  startApiServer,
}));

vi.mock("./scannerRuntime.js", () => ({
  startBackgroundScanLoop,
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
  it("boots globals, notification service, API, scanner loop, and discord bot", async () => {
    vi.resetModules();
    await import("./index.js");

    expect(initGlobals).toHaveBeenCalledTimes(1);
    expect(setNotificationService).toHaveBeenCalledTimes(1);
    expect(startApiServer).toHaveBeenCalledTimes(1);
    expect(startBackgroundScanLoop).toHaveBeenCalledTimes(1);
    expect(startDiscordBot).toHaveBeenCalledTimes(1);
  });
});
