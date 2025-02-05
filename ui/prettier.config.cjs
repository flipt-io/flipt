module.exports = {
  plugins: [
    'prettier-plugin-tailwindcss',
    '@trivago/prettier-plugin-sort-imports'
  ],
  singleQuote: true,
  trailingComma: 'none',
  importOrder: [
    '^~/app/(.*)$',
    '^~/components/(.*)$',
    '^~/types/(.*)$',
    '^~/(.*)$',
    '^[./]'
  ],
  importOrderSeparation: true,
  importOrderSortSpecifiers: true
};
