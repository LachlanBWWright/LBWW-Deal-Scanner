import { err, ok, type Result } from "neverthrow";

export interface RuntimeConfig {
  readonly scheduledMode: boolean;
  readonly scheduledDurationMs: number;
}

interface RuntimeConfigError {
  readonly type: "RuntimeConfigError";
  readonly message: string;
}

const DEFAULT_SCHEDULED_DURATION_MS = 5 * 60 * 1000;

function parseBooleanEnv(
  value: string | undefined,
): Result<boolean, RuntimeConfigError> {
  if (value === undefined || value.trim() === "") return ok(false);

  const normalized = value.trim().toLowerCase();
  if (["1", "true", "yes", "on"].includes(normalized)) return ok(true);
  if (["0", "false", "no", "off"].includes(normalized)) return ok(false);

  return err({
    type: "RuntimeConfigError",
    message:
      "SCHEDULED_SCANNER_MODE must be one of true, false, 1, 0, yes, no, on, or off",
  });
}

function parseDurationMs(
  value: string | undefined,
): Result<number, RuntimeConfigError> {
  if (value === undefined || value.trim() === "") {
    return ok(DEFAULT_SCHEDULED_DURATION_MS);
  }

  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed <= 0) {
    return err({
      type: "RuntimeConfigError",
      message: "SCHEDULED_SCANNER_DURATION_MS must be a positive integer",
    });
  }

  return ok(parsed);
}

export function readRuntimeConfig(
  env: NodeJS.ProcessEnv = process.env,
): Result<RuntimeConfig, RuntimeConfigError> {
  return parseBooleanEnv(env.SCHEDULED_SCANNER_MODE).andThen((scheduledMode) =>
    parseDurationMs(env.SCHEDULED_SCANNER_DURATION_MS).map(
      (scheduledDurationMs) => ({
        scheduledMode,
        scheduledDurationMs,
      }),
    ),
  );
}
