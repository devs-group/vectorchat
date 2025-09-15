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
            <IconArrowLeftLong class="w-4 h-4" />
            Go Back
          </Button>
          <Button @click="handleGoHome" size="lg" class="gap-2">
            <IconHome class="w-4 h-4" />
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
import IconArrowLeftLong from "@/components/icons/IconArrowLeftLong.vue";
import IconHome from "@/components/icons/IconHome.vue";

// Props for the error page
const props = defineProps<{
  error: NuxtError;
}>();

// Show detailed error in development mode
const config = useRuntimeConfig();

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
