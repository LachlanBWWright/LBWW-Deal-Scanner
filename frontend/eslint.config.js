import js from "@eslint/js";
import tseslint from "typescript-eslint";
import parser from "@typescript-eslint/parser";
import vueParser from "vue-eslint-parser";
import vue from "eslint-plugin-vue";
import globals from "globals";

export default [
  {
    ignores: ["dist", "node_modules"],
  },
  js.configs.recommended,
  ...tseslint.configs.strict,
  ...vue.configs["flat/essential"],
  {
    files: ["**/*.{ts,vue}"],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser,
        ecmaVersion: "latest",
        sourceType: "module",
        extraFileExtensions: [".vue"],
      },
      globals: {
        ...globals.browser,
      },
    },
    rules: {
      "@typescript-eslint/unified-signatures": "off",
      "no-restricted-syntax": [
        "error",
        {
          selector: "TryStatement",
          message:
            "Use neverthrow Result/ResultAsync flows instead of try/catch.",
        },
      ],
    },
  },
];
