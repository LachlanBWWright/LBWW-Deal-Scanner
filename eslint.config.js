import js from "@eslint/js";
import tseslint from "typescript-eslint";
import globals from "globals";
import eslintConfigPrettier from "eslint-config-prettier";
import neverthrowPlugin from "eslint-plugin-neverthrow";

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
    plugins: {
      neverthrow: neverthrowPlugin,
    },
    rules: {
      "@typescript-eslint/consistent-type-assertions": ["error", { "assertionStyle": "never" }],
      "no-constant-condition": ["error", { checkLoops: false }],
    }
  },
  {
    files: ["**/*.ts"],
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        project: "./tsconfig.json",
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      "@typescript-eslint/no-floating-promises": "error",
      "@typescript-eslint/no-misused-promises": "error",
      "neverthrow/must-use-result": "error",
    },
  }
];
