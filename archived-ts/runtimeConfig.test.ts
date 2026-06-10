import { describe, expect, it } from "vitest";

import { readRuntimeConfig } from "./runtimeConfig.js";

describe("runtimeConfig", () => {
  it("defaults to continuous scanner mode", () => {
    const result = readRuntimeConfig({});

    expect(result.isOk()).toBe(true);
    if (result.isOk()) {
      expect(result.value).toEqual({
        scheduledMode: false,
        scheduledDurationMs: 300000,
      });
    }
  });

  it("reads scheduled mode and duration from env", () => {
    const result = readRuntimeConfig({
      SCHEDULED_SCANNER_MODE: "true",
      SCHEDULED_SCANNER_DURATION_MS: "120000",
    });

    expect(result.isOk()).toBe(true);
    if (result.isOk()) {
      expect(result.value).toEqual({
        scheduledMode: true,
        scheduledDurationMs: 120000,
      });
    }
  });

  it("rejects invalid scheduled mode values", () => {
    const result = readRuntimeConfig({
      SCHEDULED_SCANNER_MODE: "sometimes",
    });

    expect(result.isErr()).toBe(true);
    if (result.isErr()) {
      expect(result.error.message).toContain("SCHEDULED_SCANNER_MODE");
    }
  });

  it("rejects invalid scheduled duration values", () => {
    const result = readRuntimeConfig({
      SCHEDULED_SCANNER_DURATION_MS: "0",
    });

    expect(result.isErr()).toBe(true);
    if (result.isErr()) {
      expect(result.error.message).toContain("SCHEDULED_SCANNER_DURATION_MS");
    }
  });
});
