/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        board: {
          line: "#2b2f3a",
        },
      },
    },
  },
  plugins: [],
};
