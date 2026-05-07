import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

const apiTarget =
  process.env.VITE_API_TARGET ??
  `http://127.0.0.1:${process.env.API_PORT ?? "3000"}`;

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      "/api": apiTarget,
      "/docs": apiTarget,
      "/openapi.json": apiTarget,
    },
  },
});
