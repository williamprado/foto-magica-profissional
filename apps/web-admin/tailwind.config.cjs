module.exports = {
  content: [
    "./src/**/*.{ts,tsx}",
    "./index.html",
    "../../packages/**/*.{ts,tsx}"
  ],
  theme: {
    extend: {
      colors: {
        ink: "#07070A",
        fog: "#F5F6F8",
        line: "#E6E9EF",
        accent: "#00F17C",
        accentSoft: "#D6FFE9",
        danger: "#FF4D4D"
      },
      boxShadow: {
        soft: "0 16px 48px rgba(15, 23, 42, 0.08)"
      },
      borderRadius: {
        xl2: "1.5rem"
      },
      fontFamily: {
        sans: ["Satoshi", "ui-sans-serif", "system-ui", "sans-serif"]
      }
    }
  },
  plugins: [require("@tailwindcss/forms")]
};
