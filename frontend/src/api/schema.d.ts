export interface paths {
  "/api/status": {
    parameters: {
      query?: never;
      header?: never;
      path?: never;
      cookie?: never;
    };
    /** Read runtime status */
    get: {
      parameters: {
        query?: never;
        header?: never;
        path?: never;
        cookie?: never;
      };
      requestBody?: never;
      responses: {
        /** @description Default Response */
        200: {
          headers: {
            [name: string]: unknown;
          };
          content: {
            "application/json": {
              api: {
                startedAt: string;
                port: number;
              };
              bot: {
                /** @enum {string} */
                status: "starting" | "ready" | "disabled" | "error";
                connected: boolean;
                lastReadyAt: null | string;
                lastError: null | string;
              };
              scanner: {
                /** @enum {string} */
                status: "idle" | "running" | "error";
                loopRunning: boolean;
                lastRun: null | {
                  /** @enum {string} */
                  mode: "loop" | "manual";
                  startedAt: string;
                  finishedAt: string | null;
                  durationMs: number | null;
                  error: string | null;
                };
              };
              commands: string[];
            };
          };
        };
      };
    };
    put?: never;
    post?: never;
    delete?: never;
    options?: never;
    head?: never;
    patch?: never;
    trace?: never;
  };
  "/api/commands": {
    parameters: {
      query?: never;
      header?: never;
      path?: never;
      cookie?: never;
    };
    /** List registered Discord commands */
    get: {
      parameters: {
        query?: never;
        header?: never;
        path?: never;
        cookie?: never;
      };
      requestBody?: never;
      responses: {
        /** @description Default Response */
        200: {
          headers: {
            [name: string]: unknown;
          };
          content: {
            "application/json": {
              commands: string[];
            };
          };
        };
      };
    };
    put?: never;
    post?: never;
    delete?: never;
    options?: never;
    head?: never;
    patch?: never;
    trace?: never;
  };
  "/api/scans/run": {
    parameters: {
      query?: never;
      header?: never;
      path?: never;
      cookie?: never;
    };
    get?: never;
    put?: never;
    /** Run a manual scan pass */
    post: {
      parameters: {
        query?: never;
        header?: never;
        path?: never;
        cookie?: never;
      };
      requestBody: {
        content: {
          "application/json": {
            reason?: string;
          };
        };
      };
      responses: {
        /** @description Default Response */
        200: {
          headers: {
            [name: string]: unknown;
          };
          content: {
            "application/json": {
              accepted: boolean;
              startedAt: string;
              finishedAt: string;
              durationMs: number;
              /** @enum {string} */
              mode: "manual";
              error: null | string;
            };
          };
        };
      };
    };
    delete?: never;
    options?: never;
    head?: never;
    patch?: never;
    trace?: never;
  };
  "/api/queries": {
    parameters: {
      query?: {
        type?:
          | "cashConverters"
          | "ebay"
          | "gumtree"
          | "salvos"
          | "csMarket"
          | "steamMarket"
          | "csTradeBot";
      };
      header?: never;
      path?: never;
      cookie?: never;
    };
    get: {
      parameters: {
        query?: {
          type?:
            | "cashConverters"
            | "ebay"
            | "gumtree"
            | "salvos"
            | "csMarket"
            | "steamMarket"
            | "csTradeBot";
        };
        header?: never;
        path?: never;
        cookie?: never;
      };
      requestBody?: never;
      responses: {
        /** @description Default Response */
        200: {
          headers: {
            [name: string]: unknown;
          };
          content: {
            "application/json": {
              queries: {
                type: "array";
                items: {
                  type: "object";
                  additionalProperties: false;
                  properties: {
                    type: {
                      type: "string";
                      enum: [
                        "cashConverters",
                        "ebay",
                        "gumtree",
                        "salvos",
                        "csMarket",
                        "steamMarket",
                        "csTradeBot",
                      ];
                    };
                    id: { type: "string" };
                    dmOnly: { type: "boolean" };
                    url: { type: "string" };
                    name: { type: "string" };
                    displayUrl: { type: "string" };
                    maxPrice: { type: "number" };
                    minPrice: { type: "number" };
                    maxFloat: { type: "number" };
                    requiredPhrases: { type: "string" };
                    excludePhrases: { type: "string" };
                  };
                  required: ["type", "id", "dmOnly"];
                };
              };
            };
          };
        };
      };
      put?: never;
      post?: never;
      delete?: never;
      options?: never;
      head?: never;
      patch?: never;
      trace?: never;
    };
    put: {
      parameters: {
        query?: never;
        header?: never;
        path?: never;
        cookie?: never;
      };
      requestBody: {
        content: {
          "application/json": {
            type:
              | "cashConverters"
              | "ebay"
              | "gumtree"
              | "salvos"
              | "csMarket"
              | "steamMarket"
              | "csTradeBot";
            id: string;
            payload: Record<string, unknown>;
          };
        };
      };
      responses: {
        /** @description Default Response */
        200: {
          headers: {
            [name: string]: unknown;
          };
          content: {
            "application/json": {
              success: boolean;
              query: {
                type: "string";
                enum: [
                  "cashConverters",
                  "ebay",
                  "gumtree",
                  "salvos",
                  "csMarket",
                  "steamMarket",
                  "csTradeBot",
                ];
                id: string;
                dmOnly: boolean;
                url?: string;
                name?: string;
                displayUrl?: string;
                maxPrice?: number;
                minPrice?: number;
                maxFloat?: number;
                requiredPhrases?: string;
                excludePhrases?: string;
              };
            };
          };
        };
      };
      get?: never;
      delete?: never;
      options?: never;
      head?: never;
      patch?: never;
      trace?: never;
    };
    post: {
      parameters: {
        query?: never;
        header?: never;
        path?: never;
        cookie?: never;
      };
      requestBody: {
        content: {
          "application/json": {
            type:
              | "cashConverters"
              | "ebay"
              | "gumtree"
              | "salvos"
              | "csMarket"
              | "steamMarket"
              | "csTradeBot";
            payload: Record<string, unknown>;
          };
        };
      };
      responses: {
        /** @description Default Response */
        200: {
          headers: {
            [name: string]: unknown;
          };
          content: {
            "application/json": {
              success: boolean;
              query: {
                type: "string";
                enum: [
                  "cashConverters",
                  "ebay",
                  "gumtree",
                  "salvos",
                  "csMarket",
                  "steamMarket",
                  "csTradeBot",
                ];
                id: string;
                dmOnly: boolean;
                url?: string;
                name?: string;
                displayUrl?: string;
                maxPrice?: number;
                minPrice?: number;
                maxFloat?: number;
                requiredPhrases?: string;
                excludePhrases?: string;
              };
            };
          };
        };
      };
      get?: never;
      put?: never;
      delete?: never;
      options?: never;
      head?: never;
      patch?: never;
      trace?: never;
    };
    delete: {
      parameters: {
        query?: never;
        header?: never;
        path?: never;
        cookie?: never;
      };
      requestBody: {
        content: {
          "application/json": {
            type:
              | "cashConverters"
              | "ebay"
              | "gumtree"
              | "salvos"
              | "csMarket"
              | "steamMarket"
              | "csTradeBot";
            id: string;
          };
        };
      };
      responses: {
        /** @description Default Response */
        200: {
          headers: {
            [name: string]: unknown;
          };
          content: {
            "application/json": {
              success: boolean;
            };
          };
        };
      };
      get?: never;
      put?: never;
      post?: never;
      options?: never;
      head?: never;
      patch?: never;
      trace?: never;
    };
    options?: never;
    head?: never;
    patch?: never;
    trace?: never;
  };
}
export type webhooks = Record<string, never>;
export interface components {
  schemas: never;
  responses: never;
  parameters: never;
  requestBodies: never;
  headers: never;
  pathItems: never;
}
export type $defs = Record<string, never>;
export type operations = Record<string, never>;
