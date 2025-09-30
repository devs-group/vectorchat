import tailwindcss from "@tailwindcss/vite";

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE_URL || "https://hidden-wave.podseidon.io",
      kratosPublicUrl: process.env.KRATOS_PUBLIC_URL || "http://localhost:4433",
    },
  },
  devtools: { enabled: true },
  modules: ["@nuxt/fonts", "@nuxt/scripts", "shadcn-nuxt"],
  css: ["~/assets/css/tailwind.css"],
  vite: {
    plugins: [tailwindcss()],
  },
  shadcn: {
    /**
     * Prefix for all the imported component
     */
    prefix: "",
    /**
     * Directory that the component lives in.
     * @default "./components/ui"
     */
    componentDir: "./components/ui",
  },
  ssr: false,
});
