/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./templates/**/*.html",     // matches web/ui/templates/*.html
    "./utility/js/**/*.js",      // matches web/ui/utility/js/*.js
    "./utility/css/**/*.css",     // important: includes your @apply usage
    "./web/**/*.html",
    "./web/**/*.js",
    "./web/**/*.*.css",
  ],
  safelist: ['bg-red-500', 'text-white', 'p-4', 'rounded'], // if used dynamically
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#e6f0ff',
          100: '#cce0ff',
          200: '#99c2ff',
          300: '#66a3ff',
          400: '#3385ff',
          500: '#0066ff',
          600: '#0052cc',
          700: '#003d99',
          800: '#002966',
          900: '#001433',
        }
      }
    }
  },
  plugins: [],
}

