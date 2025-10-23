// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  modules: ["@nuxt/fonts", "@nuxt/scripts", "shadcn-nuxt", "@nuxt/image"],
  css: ["@/assets/css/tailwind.css"],
  runtimeConfig: {
    vectorchatUrl: process.env.NUXT_VECTORCHAT_URL || "",
    vectorchatClientId: process.env.NUXT_VECTORCHAT_CLIENT_ID || "",
    vectorchatClientSecret: process.env.NUXT_VECTORCHAT_CLIENT_SECRET || "",
    public: {
      frontendLoginUrl: process.env.NUXT_PUBLIC_FRONTEND_LOGIN_URL || "",
      kratosPublicUrl: process.env.NUXT_PUBLIC_KRATOS_PUBLIC_URL || "",
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
