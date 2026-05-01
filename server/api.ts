import path from "node:path";
import { fileURLToPath } from "node:url";
import { existsSync } from "node:fs";

import Fastify, { type FastifyInstance, type FastifyReply, type FastifyRequest } from "fastify";
import cors from "@fastify/cors";
import swagger from "@fastify/swagger";
import swaggerUi from "@fastify/swagger-ui";
import staticPlugin from "@fastify/static";

import { getRuntimeSnapshot, setApiPort } from "./controlState.js";
import { runScanOnce } from "./scannerRuntime.js";
import { db } from "./globals/PrismaClient.js";

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
  required: [
    "accepted",
    "startedAt",
    "finishedAt",
    "durationMs",
    "mode",
    "error",
  ],
} as const;

const queryTypes = [
  "cashConverters",
  "ebay",
  "gumtree",
  "salvos",
  "csMarket",
  "steamMarket",
  "csTradeBot",
] as const;

const queryTypeEnumSchema = {
  type: "string",
  enum: queryTypes,
} as const;

const apiSecret = process.env.API_SECRET ?? "";

function getRequestSecret(request: FastifyRequest) {
  const explicitSecret = String(request.headers["x-api-secret"] ?? "").trim();
  if (explicitSecret) return explicitSecret;

  const authorization = String(request.headers.authorization ?? "").trim();
  if (authorization.startsWith("Bearer ")) {
    return authorization.slice(7).trim();
  }

  return "";
}

async function verifyApiSecret(request: FastifyRequest, reply: FastifyReply) {
  if (!request.url.startsWith("/api")) {
    return;
  }

  if (!apiSecret) {
    request.log.error("API_SECRET not configured, refusing all /api requests");
    return reply.code(500).send({ error: "Server misconfiguration" });
  }

  const provided = getRequestSecret(request);
  if (!provided || provided !== apiSecret) {
    return reply.code(401).send({ error: "Unauthorized" });
  }
}

const queryItemSchema = {
  type: "object",
  additionalProperties: false,
  properties: {
    type: queryTypeEnumSchema,
    id: { type: "string" },
    dmOnly: { type: "boolean" },
    url: { type: "string" },
    name: { type: "string" },
    displayUrl: { type: "string" },
    maxPrice: { type: "number" },
    minPrice: { type: "number" },
    maxFloat: { type: "number" },
    requiredPhrases: { type: "string" },
    excludePhrases: { type: "string" },
  },
  required: ["type", "id", "dmOnly"],
} as const;

const queryListResponseSchema = {
  type: "object",
  additionalProperties: false,
  properties: {
    queries: {
      type: "array",
      items: queryItemSchema,
    },
  },
  required: ["queries"],
} as const;

const queryCreateRequestSchema = {
  type: "object",
  additionalProperties: false,
  properties: {
    type: queryTypeEnumSchema,
    payload: { type: "object", additionalProperties: true },
  },
  required: ["type", "payload"],
} as const;

const queryMutationResponseSchema = {
  type: "object",
  additionalProperties: false,
  properties: {
    success: { type: "boolean" },
    query: queryItemSchema,
  },
  required: ["success", "query"],
} as const;

const queryDeleteRequestSchema = {
  type: "object",
  additionalProperties: false,
  properties: {
    type: queryTypeEnumSchema,
    id: { type: "string" },
  },
  required: ["type", "id"],
} as const;

const queryDeleteResponseSchema = {
  type: "object",
  additionalProperties: false,
  properties: {
    success: { type: "boolean" },
  },
  required: ["success"],
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

type QueryType = (typeof queryTypes)[number];

interface QueryItem {
  type: QueryType;
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
}

function parseBoolean(value: unknown) {
  return value === true || value === "true" || value === 1 || value === "1";
}

function normalizeNumber(value: unknown, name: string) {
  const numberValue = Number(value);
  if (Number.isNaN(numberValue)) {
    throw new Error(`${name} must be a number`);
  }
  return numberValue;
}

async function listSavedQueries(type?: QueryType) {
  const allItems: QueryItem[] = [];

  const [cashConverters, ebay, gumtree, salvos, csMarket, steamMarket, csTradeBot] =
    await Promise.all([
      db.cashConverters.findMany({ include: { query: true } }),
      db.ebay.findMany({ include: { query: true } }),
      db.gumtree.findMany({ include: { query: true } }),
      db.salvos.findMany({ include: { query: true } }),
      db.csMarket.findMany({ include: { query: true } }),
      db.steamMarket.findMany({ include: { query: true } }),
      db.csTradeBot.findMany({ include: { query: true } }),
    ]);

  allItems.push(
    ...cashConverters.map((item) => ({
      type: "cashConverters" as const,
      id: item.url,
      dmOnly: item.query?.dmOnly ?? false,
      url: item.url,
      requiredPhrases: item.requiredPhrases,
      excludePhrases: item.excludePhrases,
    })),
    ...ebay.map((item) => ({
      type: "ebay" as const,
      id: item.url,
      dmOnly: item.query?.dmOnly ?? false,
      url: item.url,
      maxPrice: item.maxPrice,
    })),
    ...gumtree.map((item) => ({
      type: "gumtree" as const,
      id: item.url,
      dmOnly: item.query?.dmOnly ?? false,
      url: item.url,
      maxPrice: item.maxPrice,
    })),
    ...salvos.map((item) => ({
      type: "salvos" as const,
      id: item.name,
      dmOnly: item.query?.dmOnly ?? false,
      name: item.name,
      minPrice: item.minPrice,
      maxPrice: item.maxPrice,
    })),
    ...csMarket.map((item) => ({
      type: "csMarket" as const,
      id: item.url,
      dmOnly: item.query?.dmOnly ?? false,
      url: item.url,
      displayUrl: item.displayUrl,
      maxPrice: item.maxPrice,
      maxFloat: item.maxFloat,
    })),
    ...steamMarket.map((item) => ({
      type: "steamMarket" as const,
      id: item.name,
      dmOnly: item.query?.dmOnly ?? false,
      name: item.name,
      displayUrl: item.displayUrl,
      maxPrice: item.maxPrice,
    })),
    ...csTradeBot.map((item) => ({
      type: "csTradeBot" as const,
      id: item.name,
      dmOnly: item.query?.dmOnly ?? false,
      name: item.name,
      minPrice: item.minFloat,
      maxFloat: item.maxFloat,
      maxPrice: item.maxPrice,
    })),
  );

  return type ? allItems.filter((item) => item.type === type) : allItems;
}

async function createSavedQuery(type: QueryType, payload: Record<string, unknown>) {
  const dmOnly = parseBoolean(payload.dmOnly);

  switch (type) {
    case "cashConverters": {
      const url = new URL(String(payload.url || "")).toString();
      const requiredPhrases = String(payload.requiredPhrases || "");
      const excludePhrases = String(payload.excludePhrases || "");
      await db.query.create({
        data: {
          dmOnly,
          cashConverters: {
            create: {
              url,
              requiredPhrases,
              excludePhrases,
            },
          },
        },
      });
      return {
        type,
        id: url,
        dmOnly,
        url,
        requiredPhrases,
        excludePhrases,
      };
    }
    case "ebay": {
      const url = new URL(String(payload.url || "")).toString();
      const maxPrice = normalizeNumber(payload.maxPrice, "maxPrice");
      await db.query.create({
        data: {
          dmOnly,
          ebay: {
            create: {
              url,
              maxPrice,
            },
          },
        },
      });
      return { type, id: url, dmOnly, url, maxPrice };
    }
    case "gumtree": {
      const url = new URL(String(payload.url || "")).toString();
      const maxPrice = normalizeNumber(payload.maxPrice, "maxPrice");
      await db.query.create({
        data: {
          dmOnly,
          gumtree: {
            create: {
              url,
              maxPrice,
            },
          },
        },
      });
      return { type, id: url, dmOnly, url, maxPrice };
    }
    case "salvos": {
      const name = String(payload.name || "").trim();
      if (!name) throw new Error("name is required");
      const minPrice = normalizeNumber(payload.minPrice, "minPrice");
      const maxPrice = normalizeNumber(payload.maxPrice, "maxPrice");
      await db.query.create({
        data: {
          dmOnly,
          salvos: {
            create: {
              name,
              minPrice,
              maxPrice,
            },
          },
        },
      });
      return { type, id: name, dmOnly, name, minPrice, maxPrice };
    }
    case "csMarket": {
      const url = new URL(String(payload.url || "")).toString();
      const displayUrl = String(payload.displayUrl || url);
      const maxPrice = normalizeNumber(payload.maxPrice, "maxPrice");
      const maxFloat = normalizeNumber(payload.maxFloat, "maxFloat");
      await db.query.create({
        data: {
          dmOnly,
          csMarket: {
            create: {
              url,
              displayUrl,
              maxPrice,
              maxFloat,
              lastPrice: 0,
            },
          },
        },
      });
      return { type, id: url, dmOnly, url, displayUrl, maxPrice, maxFloat };
    }
    case "steamMarket": {
      const name = String(payload.name || "").trim();
      if (!name) throw new Error("name is required");
      const displayUrl = String(payload.displayUrl || name);
      const maxPrice = normalizeNumber(payload.maxPrice, "maxPrice");
      await db.query.create({
        data: {
          dmOnly,
          steamMarket: {
            create: {
              name,
              displayUrl,
              maxPrice,
              lastPrice: 0,
            },
          },
        },
      });
      return { type, id: name, dmOnly, name, displayUrl, maxPrice };
    }
    case "csTradeBot": {
      const name = String(payload.name || "").trim();
      if (!name) throw new Error("name is required");
      const maxPrice = normalizeNumber(payload.maxPrice, "maxPrice");
      const minFloat = normalizeNumber(payload.minFloat, "minFloat");
      const maxFloat = normalizeNumber(payload.maxFloat, "maxFloat");
      await db.query.create({
        data: {
          dmOnly,
          csTradeBot: {
            create: {
              name,
              maxPrice,
              minFloat,
              maxFloat,
            },
          },
        },
      });
      return { type, id: name, dmOnly, name, maxPrice, minFloat, maxFloat };
    }
  }
}

async function updateSavedQuery(
  type: QueryType,
  id: string,
  payload: Record<string, unknown>,
) {
  switch (type) {
    case "cashConverters": {
      const url = new URL(String(payload.url || id)).toString();
      const requiredPhrases = String(payload.requiredPhrases || "");
      const excludePhrases = String(payload.excludePhrases || "");
      const dmOnly = parseBoolean(payload.dmOnly);
      const updated = await db.cashConverters.update({
        where: { url: id },
        data: {
          url,
          requiredPhrases,
          excludePhrases,
          query: { update: { dmOnly } },
        },
      });
      return { type, id: updated.url, dmOnly, url: updated.url, requiredPhrases, excludePhrases };
    }
    case "ebay": {
      const url = new URL(String(payload.url || id)).toString();
      const maxPrice = normalizeNumber(payload.maxPrice, "maxPrice");
      const dmOnly = parseBoolean(payload.dmOnly);
      const updated = await db.ebay.update({
        where: { url: id },
        data: {
          url,
          maxPrice,
          query: { update: { dmOnly } },
        },
      });
      return { type, id: updated.url, dmOnly, url: updated.url, maxPrice };
    }
    case "gumtree": {
      const url = new URL(String(payload.url || id)).toString();
      const maxPrice = normalizeNumber(payload.maxPrice, "maxPrice");
      const dmOnly = parseBoolean(payload.dmOnly);
      const updated = await db.gumtree.update({
        where: { url: id },
        data: {
          url,
          maxPrice,
          query: { update: { dmOnly } },
        },
      });
      return { type, id: updated.url, dmOnly, url: updated.url, maxPrice };
    }
    case "salvos": {
      const name = String(payload.name || id).trim();
      if (!name) throw new Error("name is required");
      const minPrice = normalizeNumber(payload.minPrice, "minPrice");
      const maxPrice = normalizeNumber(payload.maxPrice, "maxPrice");
      const dmOnly = parseBoolean(payload.dmOnly);
      const updated = await db.salvos.update({
        where: { name: id },
        data: {
          name,
          minPrice,
          maxPrice,
          query: { update: { dmOnly } },
        },
      });
      return { type, id: updated.name, dmOnly, name: updated.name, minPrice, maxPrice };
    }
    case "csMarket": {
      const url = new URL(String(payload.url || id)).toString();
      const displayUrl = String(payload.displayUrl || url);
      const maxPrice = normalizeNumber(payload.maxPrice, "maxPrice");
      const maxFloat = normalizeNumber(payload.maxFloat, "maxFloat");
      const dmOnly = parseBoolean(payload.dmOnly);
      const updated = await db.csMarket.update({
        where: { url: id },
        data: {
          url,
          displayUrl,
          maxPrice,
          maxFloat,
          query: { update: { dmOnly } },
        },
      });
      return {
        type,
        id: updated.url,
        dmOnly,
        url: updated.url,
        displayUrl: updated.displayUrl,
        maxPrice,
        maxFloat,
      };
    }
    case "steamMarket": {
      const name = String(payload.name || id).trim();
      if (!name) throw new Error("name is required");
      const displayUrl = String(payload.displayUrl || name);
      const maxPrice = normalizeNumber(payload.maxPrice, "maxPrice");
      const dmOnly = parseBoolean(payload.dmOnly);
      const updated = await db.steamMarket.update({
        where: { name: id },
        data: {
          name,
          displayUrl,
          maxPrice,
          query: { update: { dmOnly } },
        },
      });
      return { type, id: updated.name, dmOnly, name: updated.name, displayUrl: updated.displayUrl, maxPrice };
    }
    case "csTradeBot": {
      const name = String(payload.name || id).trim();
      if (!name) throw new Error("name is required");
      const maxPrice = normalizeNumber(payload.maxPrice, "maxPrice");
      const minFloat = normalizeNumber(payload.minFloat, "minFloat");
      const maxFloat = normalizeNumber(payload.maxFloat, "maxFloat");
      const dmOnly = parseBoolean(payload.dmOnly);
      const updated = await db.csTradeBot.update({
        where: { name: id },
        data: {
          name,
          maxPrice,
          minFloat,
          maxFloat,
          query: { update: { dmOnly } },
        },
      });
      return {
        type,
        id: updated.name,
        dmOnly,
        name: updated.name,
        maxPrice,
        minPrice: updated.minFloat,
        maxFloat: updated.maxFloat,
      };
    }
  }
}

async function deleteSavedQuery(type: QueryType, id: string) {
  switch (type) {
    case "cashConverters":
      await db.cashConverters.delete({ where: { url: id } });
      break;
    case "ebay":
      await db.ebay.delete({ where: { url: id } });
      break;
    case "gumtree":
      await db.gumtree.delete({ where: { url: id } });
      break;
    case "salvos":
      await db.salvos.delete({ where: { name: id } });
      break;
    case "csMarket":
      await db.csMarket.delete({ where: { url: id } });
      break;
    case "steamMarket":
      await db.steamMarket.delete({ where: { name: id } });
      break;
    case "csTradeBot":
      await db.csTradeBot.delete({ where: { name: id } });
      break;
  }
}

export async function buildApiServer(): Promise<FastifyInstance> {
  if (!apiSecret) {
    throw new Error("API_SECRET environment variable is required to start the API server.");
  }

  const app = Fastify({
    logger: true,
  });

  app.addHook("preHandler", verifyApiSecret);

  await app.register(cors, {
    origin: true,
    methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
    allowedHeaders: ["Content-Type", "X-API-SECRET", "Authorization"],
  });

  await app.register(swagger, {
    mode: "dynamic",
    openapi: {
      openapi: "3.1.0",
      info: {
        title: "DealScanner Control API",
        description:
          "Fastify control API for the DealScanner bot and scanner runtime.",
        version: "1.0.0",
      },
      servers: [{ url: "http://127.0.0.1:3000" }],
    },
  });

  await app.register(swaggerUi, {
    routePrefix: "/docs",
  });

  app.get(
    "/api/status",
    {
      schema: {
        tags: ["control"],
        summary: "Read runtime status",
        response: {
          200: statusResponseSchema,
        },
      },
    },
    async () => getRuntimeSnapshot(),
  );

  app.get(
    "/api/commands",
    {
      schema: {
        tags: ["control"],
        summary: "List registered Discord commands",
        response: {
          200: commandListResponseSchema,
        },
      },
    },
    async () => {
      const snapshot = getRuntimeSnapshot();
      return { commands: snapshot.commands };
    },
  );

  app.post(
    "/api/scans/run",
    {
      schema: {
        tags: ["control"],
        summary: "Run a manual scan pass",
        body: runScanRequestSchema,
        response: {
          200: runScanResponseSchema,
        },
      },
    },
    async () => {
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
    },
  );

  app.get(
    "/api/queries",
    {
      schema: {
        tags: ["control"],
        summary: "List saved queries",
        querystring: {
          type: "object",
          additionalProperties: false,
          properties: {
            type: queryTypeEnumSchema,
          },
        },
        response: {
          200: queryListResponseSchema,
        },
      },
    },
    async (request: FastifyRequest<{ Querystring: { type?: QueryType } }>) => {
      const queryType = request.query.type;
      const queries = await listSavedQueries(queryType);
      return { queries };
    },
  );

  app.post(
    "/api/queries",
    {
      schema: {
        tags: ["control"],
        summary: "Create a saved query",
        body: queryCreateRequestSchema,
        response: {
          200: queryMutationResponseSchema,
        },
      },
    },
    async (request: FastifyRequest<{ Body: { type: QueryType; payload: Record<string, unknown> } }>) => {
      const body = request.body;
      const query = await createSavedQuery(body.type, body.payload);
      return { success: true, query };
    },
  );

  app.put(
    "/api/queries",
    {
      schema: {
        tags: ["control"],
        summary: "Update a saved query",
        body: {
          type: "object",
          additionalProperties: false,
          properties: {
            type: queryTypeEnumSchema,
            id: { type: "string" },
            payload: { type: "object", additionalProperties: true },
          },
          required: ["type", "id", "payload"],
        },
        response: {
          200: queryMutationResponseSchema,
        },
      },
    },
    async (request: FastifyRequest<{ Body: { type: QueryType; id: string; payload: Record<string, unknown> } }>) => {
      const body = request.body;
      const query = await updateSavedQuery(body.type, body.id, body.payload);
      return { success: true, query };
    },
  );

  app.delete(
    "/api/queries",
    {
      schema: {
        tags: ["control"],
        summary: "Delete a saved query",
        body: queryDeleteRequestSchema,
        response: {
          200: queryDeleteResponseSchema,
        },
      },
    },
    async (request: FastifyRequest<{ Body: { type: QueryType; id: string } }>) => {
      const body = request.body;
      await deleteSavedQuery(body.type, body.id);
      return { success: true };
    },
  );

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
