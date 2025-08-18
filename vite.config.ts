import { defineConfig } from "vite";
import nodeExternals from "vite-plugin-node-externals";

export default defineConfig({
  plugins: [nodeExternals()],
  build: {
    target: "node18",
    outDir: "build",
    rollupOptions: {
      input: "./index.ts",
      output: {
        entryFileNames: "index.js",
        format: "esm",
      },
    },
    minify: false,
    sourcemap: true,
  },
});
