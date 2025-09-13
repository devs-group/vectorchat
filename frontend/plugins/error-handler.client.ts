import { useErrorHandler } from "~/composables/useErrorHandler";

export default defineNuxtPlugin(() => {
  // Make error handler available globally
  const errorHandler = useErrorHandler();

  // Provide global error handling for unhandled promise rejections
  // Only run on client side (the .client.ts suffix already ensures this)
  window.addEventListener("unhandledrejection", (event) => {
    console.error("Unhandled promise rejection:", event.reason);
    errorHandler.showError(event.reason);
  });

  return {
    provide: {
      errorHandler,
    },
  };
});
