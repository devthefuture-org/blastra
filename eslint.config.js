import js from "@eslint/js"
import react from "eslint-plugin-react"
import reactHooks from "eslint-plugin-react-hooks"
import prettier from "eslint-plugin-prettier"
import * as parser from "@babel/eslint-parser"

export default [
  {
    ignores: [
      "**/dist/**",
      "**/build/**",
      "tests/**",
      "**/node_modules/**",
      "**/.next/**",
      "**/coverage/**",
      "**/tmp/**",
      "**/vendor/**",
      "packages/*/dist/**",
      "packages/*/build/**",
    ],
  },
  js.configs.recommended,
  {
    files: ["**/*.{js,jsx,ts,tsx}"],
    plugins: { react, "react-hooks": reactHooks, prettier },
    languageOptions: {
      parser,
      ecmaVersion: 2024,
      sourceType: "module",
      parserOptions: {
        ecmaFeatures: {
          jsx: true,
        },
        requireConfigFile: false,
        babelOptions: {
          presets: ["@babel/preset-react"],
        },
      },
      globals: {
        window: true,
        document: true,
        console: true,
        fetch: true,
        process: true,
        setTimeout: true,
        clearTimeout: true,
        FormData: true,
        history: true,
        location: true,
        addEventListener: true,
        removeEventListener: true,
        dispatchEvent: true,
        Event: true,
        performance: true,
        MessageChannel: true,
        AbortController: true,
        queueMicrotask: true,
        matchMedia: true,
        __REACT_DEVTOOLS_GLOBAL_HOOK__: true,
        reportError: true,
        setImmediate: true,
      },
    },
    settings: {
      react: {
        version: "detect",
        "jsx-runtime": true,
      },
    },
    rules: {
      "prettier/prettier": "error",
      "react/react-in-jsx-scope": "off",
      "react/prop-types": "off",
      "react/jsx-uses-vars": "error",
      "react/jsx-uses-react": "error",
      "no-unused-vars": [
        "error",
        {
          varsIgnorePattern: "^_",
          argsIgnorePattern: "^_",
          destructuredArrayIgnorePattern: "^_",
          caughtErrors: "none",
        },
      ],
      "no-empty": ["error", { allowEmptyCatch: true }],
      "no-case-declarations": "off",
    },
  },
  {
    files: ["**/*.cjs"],
    languageOptions: {
      sourceType: "commonjs",
      globals: {
        __dirname: true,
        process: true,
        console: true,
        require: true,
        module: true,
        exports: true,
      },
    },
  },
]
