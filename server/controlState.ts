export type BotStatus = "starting" | "ready" | "disabled" | "error";
export type ScanStatus = "idle" | "running" | "error";

export interface ScanRunRecord {
  mode: "loop" | "manual";
  startedAt: string;
  finishedAt: string | null;
  durationMs: number | null;
  error: string | null;
}

export interface RuntimeSnapshot {
  api: {
    startedAt: string;
    port: number;
  };
  bot: {
    status: BotStatus;
    connected: boolean;
    lastReadyAt: string | null;
    lastError: string | null;
  };
  scanner: {
    status: ScanStatus;
    loopRunning: boolean;
    lastRun: ScanRunRecord | null;
    recentResults: SearchResultRecord[];
  };
  commands: string[];
}

export interface SearchResultRecord {
  source: string;
  title: string;
  url: string;
  price: number | null;
  imageUrl: string | null;
  queryType: string | null;
  queryId: string | null;
  foundAt: string;
}

const state: RuntimeSnapshot = {
  api: {
    startedAt: new Date().toISOString(),
    port: Number(process.env.API_PORT ?? 3000),
  },
  bot: {
    status: "starting",
    connected: false,
    lastReadyAt: null,
    lastError: null,
  },
  scanner: {
    status: "idle",
    loopRunning: false,
    lastRun: null,
    recentResults: [],
  },
  commands: [],
};

const MAX_RECENT_RESULTS = 100;

export function setApiPort(port: number) {
  state.api.port = port;
}

export function setBotStatus(status: BotStatus, lastError: string | null = null) {
  state.bot.status = status;
  state.bot.connected = status === "ready";
  state.bot.lastError = lastError;
  if (status === "ready") {
    state.bot.lastReadyAt = new Date().toISOString();
  }
}

export function setCommandNames(commands: string[]) {
  state.commands = [...commands].sort((left, right) => left.localeCompare(right));
}

export function beginScanLoop() {
  state.scanner.loopRunning = true;
}

export function endScanLoop() {
  state.scanner.loopRunning = false;
}

export function beginScanRun(mode: ScanRunRecord["mode"]): ScanRunRecord {
  const startedAt = new Date().toISOString();
  state.scanner.status = "running";
  const record: ScanRunRecord = {
    mode,
    startedAt,
    finishedAt: null,
    durationMs: null,
    error: null,
  };
  state.scanner.lastRun = record;
  return record;
}

export function finishScanRun(record: ScanRunRecord, error: string | null = null) {
  const finishedAt = new Date().toISOString();
  const durationMs = new Date(finishedAt).getTime() - new Date(record.startedAt).getTime();
  record.finishedAt = finishedAt;
  record.durationMs = durationMs;
  record.error = error;
  state.scanner.status = error ? "error" : "idle";
}

export function recordSearchResults(results: SearchResultRecord[]) {
  if (results.length === 0) return;
  state.scanner.recentResults = [
    ...results,
    ...state.scanner.recentResults,
  ].slice(0, MAX_RECENT_RESULTS);
}

export function getRuntimeSnapshot(): RuntimeSnapshot {
  return {
    api: { ...state.api },
    bot: { ...state.bot },
    scanner: {
      ...state.scanner,
      lastRun: state.scanner.lastRun ? { ...state.scanner.lastRun } : null,
      recentResults: state.scanner.recentResults.map((result) => ({ ...result })),
    },
    commands: [...state.commands],
  };
}
