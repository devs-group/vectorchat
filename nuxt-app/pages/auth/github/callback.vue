<script setup>
import { useRouter, useRoute } from "vue-router";

const router = useRouter();
const route = useRoute();
const apiService = useApiService();

const queryParams = new URLSearchParams(route.query).toString();
const {
  execute: githubAuthCallback,
  isLoading,
  data: githubAuthCallbackResponse,
} = apiService.githubAuthCallback(queryParams);

const handleCallback = async () => {
  try {
    const code = route.query.code;
    if (!code) {
      throw new Error("No authorization code received from GitHub");
    }

    // Call your API to exchange the code for an access token
    await githubAuthCallback();
    if (githubAuthCallbackResponse.value) {
      console.log(githubAuthCallbackResponse.value);
    }

    // Redirect to dashboard after successful authentication
    router.push("/chat");
  } catch (error) {
    console.error("GitHub authentication error:", error);
    router.push("/auth/login?error=github_auth_failed");
  }
};

// Execute the callback handler when the component is mounted
onMounted(() => {
  handleCallback();
});
</script>

<template>
  <div class="github-callback">
    <div class="flex flex-col items-center justify-center min-h-screen p-4">
      <div
        class="w-full max-w-md p-8 space-y-4 bg-white rounded-lg shadow dark:bg-gray-800"
      >
        <h1
          class="text-2xl font-bold text-center text-gray-900 dark:text-white"
        >
          Authenticating with GitHub
        </h1>
        <div class="flex justify-center mt-4">
          <div
            class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"
          ></div>
        </div>
        <p class="text-center text-gray-600 dark:text-gray-300">
          Please wait while we complete your authentication process...
        </p>
      </div>
    </div>
  </div>
</template>
