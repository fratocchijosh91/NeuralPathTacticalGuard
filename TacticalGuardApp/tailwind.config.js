/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./app/**/*.{js,jsx,ts,tsx}", "./components/**/*.{js,jsx,ts,tsx}"],
  presets: [require("nativewind/preset")],
  theme: {
    extend: {
      colors: {
        cyber: {
          bg: "#030806",
          panel: "#0a120e",
          border: "#1a3d22",
          neon: "#39ff14",
          "neon-dim": "#6fdc5c",
          muted: "#4a6b52",
        },
      },
      fontFamily: {
        mono: ["Menlo", "Monaco", "Courier New", "monospace"],
      },
    },
  },
  plugins: [],
};
