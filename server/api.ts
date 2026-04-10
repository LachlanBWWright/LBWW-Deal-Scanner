import path from "node:path";
import { fileURLToPath } from "node:url";
import { existsSync } from "node:fs";

import Fastify, { type FastifyInstance } from "fastify";
import cors from "@fastify/cors";
import swagger from "@fastify/swagger";
import swaggerUi from "@fastify/swagger-ui";
import staticPlugin from "@fastify/static";

import { getRuntimeSnapshot, setApiPort } from "./controlState.js";
import { runScanOnce } from "./scannerRuntime.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const distDirCandidates = [
  path.resolve(__dirname, "../frontend/dist"),
  path.resolve(__dirname, "../../frontend/dist"),
];
const distDir = distDirCandidates.find(existsSync) ?? distDirCandidates[0];

const statusResponseSchema = {
  type: "object",
  additionalProperties: false,
  properties: {
    api: {
      type: "object",
      additionalProperties: false,
      properties: {
        startedAt: { type: "string" },
        port: { type: "number" },
      },
      required: ["startedAt", "port"],
    },
    bot: {
      type: "object",
      additionalProperties: false,
      properties: {
        status: {
          type: "string",
          enum: ["starting", "ready", "disabled", "error"],
        },
        connected: { type: "boolean" },
        lastReadyAt: { type: ["string", "null"] },
        lastError: { type: ["string", "null"] },
      },
      required: ["status", "connected", "lastReadyAt", "lastError"],
    },
    scanner: {
      type: "object",
      additionalProperties: false,
      properties: {
        status: {
          type: "string",
          enum: ["idle", "running", "error"],
        },
        loopRunning: { type: "boolean" },
        lastRun: {
          anyOf: [
            { type: "null" },
            {
              type: "object",
              additionalProperties: false,
              properties: {
                mode: { type: "string", enum: ["loop", "manual"] },
                startedAt: { type: "string" },
                finishedAt: { type: ["string", "null"] },
                durationMs: { type: ["number", "null"] },
                error: { type: ["string", "null"] },
              },
              required: [
                "mode",
                "startedAt",
                "finishedAt",
                "durationMs",
                "error",
              ],
            },
          ],
        },
      },
      required: ["status", "loopRunning", "lastRun"],
    },
    commands: {
      type: "array",
      items: { type: "string" },
    },
  },
  required: ["api", "bot", "scanner", "commands"],
} as const;

const runScanRequestSchema = {
  type: "object",
  additionalProperties: false,
  properties: {
    reason: { type: "string" },
  },
} as const;

const runScanResponseSchema = {
  type: "object",
  additionalProperties: false,
  properties: {
    accepted: { type: "boolean" },
    startedAt: { type: "string" },
    finishedAt: { type: "string" },
    durationMs: { type: "number" },
    mode: { type: "string", enum: ["manual"] },
    error: { type: ["string", "null"] },
  },
  required: ["accepted", "startedAt", "finishedAt", "durationMs", "mode", "error"],
} as const;

const commandListResponseSchema = {
  type: "object",
  additionalProperties: false,
  properties: {
    commands: {
      type: "array",
      items: { type: "string" },
    },
  },
  required: ["commands"],
} as const;

export async function buildApiServer(): Promise<FastifyInstance> {
  const app = Fastify({
    logger: true,
  });

  await app.register(cors, {
    origin: true,
  });

  await app.register(swagger, {
    mode: "dynamic",
    openapi: {
      openapi: "3.1.0",
      info: {
        title: "DealScanner Control API",
        description: "Fastify control API for the DealScanner bot and scanner runtime.",
        version: "1.0.0",
      },
      servers: [{ url: "http://127.0.0.1:3000" }],
    },
  });

  await app.register(swaggerUi, {
    routePrefix: "/docs",
  });

  app.get("/api/status", {
    schema: {
      tags: ["control"],
      summary: "Read runtime status",
      response: {
        200: statusResponseSchema,
      },
    },
  }, async () => getRuntimeSnapshot());

  app.get("/api/commands", {
    schema: {
      tags: ["control"],
      summary: "List registered Discord commands",
      response: {
        200: commandListResponseSchema,
      },
    },
  }, async () => {
    const snapshot = getRuntimeSnapshot();
    return { commands: snapshot.commands };
  });

  app.post("/api/scans/run", {
    schema: {
      tags: ["control"],
      summary: "Run a manual scan pass",
      body: runScanRequestSchema,
      response: {
        200: runScanResponseSchema,
      },
    },
  }, async () => {
    const record = await runScanOnce();
    if (!record.finishedAt || record.durationMs === null) {
      throw new Error("Scan record was not finalized correctly.");
    }

    return {
      accepted: true,
      startedAt: record.startedAt,
      finishedAt: record.finishedAt,
      durationMs: record.durationMs,
      mode: record.mode,
      error: record.error,
    };
  });

  if (existsSync(distDir)) {
    await app.register(staticPlugin, {
      root: distDir,
      prefix: "/",
    });

    app.setNotFoundHandler((request, reply) => {
      if (
        request.url.startsWith("/api") ||
        request.url.startsWith("/docs") ||
        request.url.startsWith("/openapi")
      ) {
        void reply.code(404).send({ message: "Not Found" });
        return;
      }

      void reply.type("text/html").sendFile("index.html");
    });
  }

  return app;
}

export async function startApiServer(app = buildApiServer()) {
  const server = await app;
  const port = Number(process.env.API_PORT ?? 3000);
  const host = process.env.API_HOST ?? "0.0.0.0";
  setApiPort(port);
  await server.listen({ port, host });
  return { app: server, port, host };
}
