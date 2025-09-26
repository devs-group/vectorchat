import { defineConfig } from "vite";

export default defineConfig({
  server: {
    open: true,
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
    sourcemap: false,
    minify: "esbuild",
    rollupOptions: {
      input: "vectorchat-glass-widget.js",
      output: {
        format: "iife",
        entryFileNames: "vectorchat-glass-widget.min.js",
        assetFileNames: "[name][extname]",
      },
    },
  },
});
