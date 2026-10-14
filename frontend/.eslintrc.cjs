module.exports = {
  root: true,
  env: {
    browser: true,
    es2021: true,
    node: true,
  },
  parser: "vue-eslint-parser",
  parserOptions: {
    parser: "@typescript-eslint/parser",
    ecmaVersion: "latest",
    sourceType: "module",
    extraFileExtensions: [".vue"],
  },
  plugins: ["vue", "@typescript-eslint"],
  extends: [
    "eslint:recommended",
    // [CUSTOM] eslint-plugin-vue 10 把 vue3-essential 更名为 essential（Vue 3 为默认），见 CUSTOMIZATIONS.md。
    "plugin:vue/essential",
    "plugin:@typescript-eslint/recommended",
  ],
  rules: {
    "no-constant-condition": "off",
    "no-mixed-spaces-and-tabs": "off",
    "no-useless-escape": "off",
    "no-unused-vars": "off",
    "@typescript-eslint/no-unused-vars": [
      "warn",
      { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
    ],
    "@typescript-eslint/ban-types": "off",
    // [CUSTOM] typescript-eslint 8 把 ban-types 拆成以下规则并加入 recommended；保持原有 off 语义，见 CUSTOMIZATIONS.md。
    "@typescript-eslint/no-empty-object-type": "off",
    "@typescript-eslint/no-unsafe-function-type": "off",
    "@typescript-eslint/no-unused-expressions": "off",
    "@typescript-eslint/ban-ts-comment": "off",
    "@typescript-eslint/no-explicit-any": "off",
    "vue/multi-word-component-names": "off",
    "vue/no-use-v-if-with-v-for": "off",
  },
};
