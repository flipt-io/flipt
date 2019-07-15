// https://eslint.org/docs/user-guide/configuring

module.exports = {
  root: true,
  parserOptions: {
    parser: "babel-eslint",
    allowImportExportEverywhere: true,
    ecmaVersion: 2017,
    sourceType: 'module',
  },
  env: {
    browser: true,
    node: true
  },
  extends: [
    "plugin:vue/base",
    'plugin:vue/recommended',
    'eslint:recommended',
    'plugin:prettier/recommended',
    'prettier/vue'
  ],
  // required to lint *.vue files
  plugins: [
    'vue'
  ],
  // add your custom rules here
  rules: {
    'no-console': 'off',
    // allow async-await
    'generator-star-spacing': 'off',
    // allow debugger during development
    'no-debugger': process.env.NODE_ENV === 'production' ? 'error' : 'off',
    'vue/max-attributes-per-line': ['error', {
      'singleline': 5,
      'multiline': {
        'max': 5,
        'allowFirstLine': true
      }
    }],
    'vue/require-default-prop': 'off',
    'vue/component-name-in-template-casing': ['error', 'PascalCase', {
      'ignores': ['draggable']
    }]
  }
}
