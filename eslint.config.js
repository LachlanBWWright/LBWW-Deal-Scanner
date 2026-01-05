import js from "@eslint/js";
import tseslint from "typescript-eslint";
import globals from "globals";
import eslintConfigPrettier from "eslint-config-prettier";

export default [
  { ignores: ["dist", "build", "coverage", ".cache", "dev.db", "node_modules"] },
  js.configs.recommended,
  ...tseslint.configs.strict,
  ...tseslint.configs.stylistic,
  eslintConfigPrettier,
  {
    languageOptions: {
      globals: {
        ...globals.node,
        ...globals.browser,
      },
    },
    rules: {
      "@typescript-eslint/consistent-type-assertions": ["error", { "assertionStyle": "never" }],
    }
  }
];
