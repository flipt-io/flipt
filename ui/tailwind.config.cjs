/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    nightwind: {
      colors: {
        white: 'gray.950',
        black: 'gray.100',
        violet: {
          400: '#8B5CF6',
          300: '#A78BFA',
          200: '#C4B5FD'
        }
      },
      colorScale: {
        preset: 'reduced'
      }
    },
    extend: {}
  },
  darkMode: 'class',
  plugins: [
    require('@tailwindcss/forms'),
    require('nightwind'),
    require('tailwindcss-bg-patterns')
  ]
};
