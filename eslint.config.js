import js from "@eslint/js";
import tseslint from "typescript-eslint";
import globals from "globals";
import eslintConfigPrettier from "eslint-config-prettier";

export default tseslint.config(
  { ignores: ["dist", "build", "coverage", ".cache", "dev.db", "node_modules"] },
  js.configs.recommended,
  ...tseslint.configs.strict,
  eslintConfigPrettier,
  {
    languageOptions: {
      globals: {
        ...globals.node,
        ...globals.browser,
      },
    },
  }
);
