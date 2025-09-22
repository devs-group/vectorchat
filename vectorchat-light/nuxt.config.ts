// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  modules: ["@nuxt/fonts", "@nuxt/scripts", "shadcn-nuxt"],
  css: ["@/assets/css/tailwind.css"],
  components: [
    {
      path: "./app/components",
      extensions: ["vue"],
      pathPrefix: false,
      ignore: ["**/ui/**"],
    },
  ],
  postcss: {
    plugins: {
      "@tailwindcss/postcss": {},
      autoprefixer: {},
    },
  },
  shadcn: {
    prefix: "",
    componentDir: "./app/components/ui",
  },
});
