import type { Config } from "tailwindcss";

export default {
  darkMode: ["class"],
  content: ["./app/**/*.{js,ts,vue}", "./nuxt.config.{js,ts}"],
  theme: {
    extend: {},
  },
  plugins: [],
} satisfies Config;
