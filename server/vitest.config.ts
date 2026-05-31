import { defineConfig } from "vite";
import { configDefaults } from "vitest/config";

export default defineConfig({
  test: {
    exclude: [...configDefaults.exclude, "build/*"],
    fileParallelism: false,
    testTimeout: 25_000,
  },
});
