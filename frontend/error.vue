<template>
  <div
    class="min-h-screen flex items-center justify-center bg-gradient-to-br from-background via-background to-muted"
  >
    <div class="max-w-md w-full px-6 py-8">
      <div class="text-center space-y-6">
        <!-- 404 Icon/Illustration -->
        <div class="relative">
          <div class="text-9xl font-bold text-muted-foreground/20 select-none">
            404
          </div>
        </div>

        <!-- Error Message -->
        <div class="space-y-2">
          <h1 class="text-3xl font-bold tracking-tight">
            {{
              error?.statusCode === 404
                ? "Page not found"
                : "Something went wrong"
            }}
          </h1>
          <p class="text-muted-foreground text-lg">
            {{
              error?.statusCode === 404
                ? "The page you're looking for doesn't exist or has been moved."
                : error?.statusMessage ||
                  "An unexpected error occurred. Please try again."
            }}
          </p>
        </div>

        <!-- Action Buttons -->
        <div class="flex flex-col sm:flex-row gap-3 justify-center pt-4">
          <Button
            @click="handleGoBack"
            variant="outline"
            size="lg"
            class="gap-2"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M10 19l-7-7m0 0l7-7m-7 7h18"
              />
            </svg>
            Go Back
          </Button>
          <Button @click="handleGoHome" size="lg" class="gap-2">
            <svg
              class="w-4 h-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
              />
            </svg>
            Go to Home
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { NuxtError } from "#app";
import { Button } from "~/components/ui/button";

// Props for the error page
const props = defineProps<{
  error: NuxtError;
}>();

// Show detailed error in development mode
const config = useRuntimeConfig();
const showError =
  config.public.nodeEnv === "development" || import.meta.env.DEV;

// Handle navigation
const handleGoBack = () => {
  if (window.history.length > 1) {
    window.history.back();
  } else {
    navigateTo("/chat");
  }
};

const handleGoHome = () => {
  // Clear the error and navigate to chat
  clearError({ redirect: "/chat" });
};

// Set page meta
useHead({
  title: `${props.error?.statusCode || "Error"} - Page Not Found`,
});
</script>

<style scoped>
/* Add subtle animation to the 404 text */
@keyframes float {
  0%,
  100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-10px);
  }
}

.text-9xl {
  animation: float 3s ease-in-out infinite;
}
</style>
