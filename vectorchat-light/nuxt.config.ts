import tailwindcss from "@tailwindcss/vite";

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  modules: [
    "@nuxt/fonts",
    "@nuxt/scripts",
    "shadcn-nuxt",
    "@nuxtjs/tailwindcss",
  ],
  css: ["@/assets/css/tailwind.css"],
  components: [
    {
      path: "./app/components",
      extensions: ["vue"],
      pathPrefix: false,
      ignore: ["**/ui/**"],
    },
  ],
  vite: {
    plugins: [tailwindcss()],
  },
  shadcn: {
    prefix: "",
    componentDir: "./app/components/ui",
  },
});