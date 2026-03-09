import { defineConfig } from "vite";
import nodeExternals from "vite-plugin-node-externals";

export default defineConfig({
  plugins: [nodeExternals()],
  // ensure that `process.env` references are left intact during bundling
  // dotenv will populate the real `process.env` at runtime.  By default
  // esbuild replaces `process.env` with an empty object, which is what
  // caused our configuration values to vanish after building.
  define: {
    // this tells Vite to substitute the identifier with itself instead
    // of statically inlining an empty object.
    "process.env": "process.env",
  },
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
