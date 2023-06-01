/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    nightwind: {
      colorScale: {
        preset: 'reduced'
      }
    },
    extend: {}
  },
  darkMode: 'class',
  plugins: [require('@tailwindcss/forms'), require('nightwind')]
};
