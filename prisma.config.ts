import "dotenv/config";
import { defineConfig } from "prisma/config";

export default defineConfig({
  schema: "prisma/schema/schema.prisma",
  migrations: {
    path: "prisma/migrations",
  },
  datasource: {
    url: console.log(process.env.DATABASE_URL) ?? "file:./prisma/dev.db",
  },
});
