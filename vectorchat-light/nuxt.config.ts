// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  modules: ["@nuxt/fonts", "@nuxt/scripts", "shadcn-nuxt", "@nuxt/image"],
  css: ["@/assets/css/tailwind.css"],
  runtimeConfig: {
    vectorchatUrl: "", // env:NUXT_VECTORCHAT_URL
    vectorchatApiKey: "", // env:NUXT_VECTORCHAT_API_KEY
    public: {
      vectorchatUrl: "", // env:NUXT_PUBLIC_VECTORCHAT_URL
      frontendLoginUrl:
        process.env.NUXT_PUBLIC_FRONTEND_LOGIN_URL || "http://localhost:3000/login",
      kratosPublicUrl:
        process.env.NUXT_PUBLIC_KRATOS_PUBLIC_URL || "http://localhost:4433",
    },
  },
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
