/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./web/ui/templates/**/*.html",      // Path from project root to my HTML files
    "./web/ui/utility/js/**/*.js",    // Path from project root to my JS files
  ],
   safelist: [
    'text-primary-500',
    'text-primary-700',
    'bg-primary-600',
    'hover:bg-primary-700',
    'border-primary-300',
    // Add other primary color classes you're using
  ],
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
};
