import { mkdir, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

import openapiTS, { astToString } from "openapi-typescript";

import { buildApiServer } from "../api.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const repoRoot = path.resolve(__dirname, "..");
const apiDir = path.resolve(repoRoot, "../frontend/src/api");
const openApiPath = path.resolve(apiDir, "openapi.json");
const typesPath = path.resolve(apiDir, "schema.d.ts");

async function run() {
  const app = await buildApiServer({ requireApiSecret: false });
  await app.ready();

  const spec = app.swagger();
  await mkdir(apiDir, { recursive: true });
  await writeFile(openApiPath, `${JSON.stringify(spec, null, 2)}\n`, "utf8");

  const types = await openapiTS(pathToFileURL(openApiPath));
  await writeFile(typesPath, `${astToString(types)}\n`, "utf8");

  await app.close();
}

run().catch((error: unknown) => {
  console.error(error);
  process.exitCode = 1;
});
