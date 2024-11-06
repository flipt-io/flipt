/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        white: 'rgb(var(--white)/<alpha-value>)',
        black: 'rgb(var(--black)/<alpha-value>)',
        gray: {
          50: 'rgb(var(--gray-50)/<alpha-value>)',
          100: 'rgb(var(--gray-100)/<alpha-value>)',
          200: 'rgb(var(--gray-200)/<alpha-value>)',
          300: 'rgb(var(--gray-300)/<alpha-value>)',
          400: 'rgb(var(--gray-400)/<alpha-value>)',
          500: 'rgb(var(--gray-500)/<alpha-value>)',
          600: 'rgb(var(--gray-600)/<alpha-value>)',
          700: 'rgb(var(--gray-700)/<alpha-value>)',
          800: 'rgb(var(--gray-800)/<alpha-value>)',
          900: 'rgb(var(--gray-900)/<alpha-value>)',
          950: 'rgb(var(--gray-950)/<alpha-value>)',

          // tailwindcss-bg-patterns doesn't support dynamic alpha colors.
          // this is a workaround.
          solid: 'rgb(var(--white))',
          solid50: 'rgb(var(--gray-50))',
          solid100: 'rgb(var(--gray-100))'
        },
        violet: {
          600: 'rgb(var(--violet-600)/<alpha-value>)',
          400: 'rgb(var(--violet-400)/<alpha-value>)',
          300: 'rgb(var(--violet-300)/<alpha-value>)',
          200: 'rgb(var(--violet-200)/<alpha-value>)'
        },
        green: {
          800: 'rgb(var(--green-800)/<alpha-value>)',
          600: 'rgb(var(--green-600)/<alpha-value>)',
          500: 'rgb(var(--green-500)/<alpha-value>)',
          400: 'rgb(var(--green-400)/<alpha-value>)',
          100: 'rgb(var(--green-100)/<alpha-value>)',
          50: 'rgb(var(--green-50)/<alpha-value>)'
        },
        red: {
          800: 'rgb(var(--red-800)/<alpha-value>)',
          700: 'rgb(var(--red-700)/<alpha-value>)',
          600: 'rgb(var(--red-600)/<alpha-value>)',
          500: 'rgb(var(--red-500)/<alpha-value>)',
          400: 'rgb(var(--red-400)/<alpha-value>)',
          100: 'rgb(var(--red-100)/<alpha-value>)',
          50: 'rgb(var(--red-50)/<alpha-value>)'
        }
      },
      colorScale: {
        preset: 'reduced'
      }
    }
  },
  darkMode: 'class',
  plugins: [require('@tailwindcss/forms'), require('tailwindcss-bg-patterns')]
};
