/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
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
          solid: 'rgb(var(--white))',
          solid50: 'rgb(var(--gray-50))',
          solid100: 'rgb(var(--gray-100))'
        },
        violet: {
          200: 'rgb(var(--violet-200)/<alpha-value>)',
          300: 'rgb(var(--violet-300)/<alpha-value>)',
          400: 'rgb(var(--violet-400)/<alpha-value>)',
          600: 'rgb(var(--violet-600)/<alpha-value>)'
        },
        green: {
          50: 'rgb(var(--green-50)/<alpha-value>)',
          100: 'rgb(var(--green-100)/<alpha-value>)',
          400: 'rgb(var(--green-400)/<alpha-value>)',
          500: 'rgb(var(--green-500)/<alpha-value>)',
          600: 'rgb(var(--green-600)/<alpha-value>)',
          800: 'rgb(var(--green-800)/<alpha-value>)'
        },
        red: {
          50: 'rgb(var(--red-50)/<alpha-value>)',
          100: 'rgb(var(--red-100)/<alpha-value>)',
          400: 'rgb(var(--red-400)/<alpha-value>)',
          500: 'rgb(var(--red-500)/<alpha-value>)',
          600: 'rgb(var(--red-600)/<alpha-value>)',
          700: 'rgb(var(--red-700)/<alpha-value>)',
          800: 'rgb(var(--red-800)/<alpha-value>)'
        },
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        card: {
          DEFAULT: 'hsl(var(--card))',
          foreground: 'hsl(var(--card-foreground))'
        },
        popover: {
          DEFAULT: 'hsl(var(--popover))',
          foreground: 'hsl(var(--popover-foreground))'
        },
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))'
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))'
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))'
        },
        accent: {
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))'
        },
        destructive: {
          DEFAULT: 'hsl(var(--destructive))',
          foreground: 'hsl(var(--destructive-foreground))'
        },
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        chart: {
          1: 'hsl(var(--chart-1))',
          2: 'hsl(var(--chart-2))',
          3: 'hsl(var(--chart-3))',
          4: 'hsl(var(--chart-4))',
          5: 'hsl(var(--chart-5))'
        }
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)'
      }
    }
  },
  darkMode: ['class'],
  plugins: [
    require('@tailwindcss/forms'),
    require('tailwindcss-bg-patterns'),
    require('tailwindcss-animate')
  ]
};
