/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js}'],
  theme: {
    extend: {
      colors: {
        ins: '#e6ffed',
        del: '#ffebe9',
        rep: '#fff8c5',
      },
    },
  },
  plugins: [],
}
