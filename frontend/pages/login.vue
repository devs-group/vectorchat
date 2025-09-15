<template>
  <div class="flex min-h-[calc(100vh-3.5rem)] items-center justify-center">
    <div
      class="mx-auto flex w-full flex-col justify-center space-y-6 sm:w-[350px]"
    >
      <div class="flex flex-col space-y-2 text-center">
        <h1 class="text-2xl font-semibold tracking-tight">VectorChat</h1>
        <p class="text-sm text-muted-foreground">
          Sign in to your account to continue
        </p>
      </div>
      <div class="grid gap-6">
        <Button variant="outline" class="w-full" @click="handleGithubLogin">
          <IconGithub
            class="mr-2 h-4 w-4"
            aria-hidden="true"
            focusable="false"
            data-prefix="fab"
            data-icon="github"
            role="img"
          />
          Continue with GitHub
        </Button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import IconGithub from "@/components/icons/IconGithub.vue";

const apiService = useApiService();

const handleGithubLogin = async () => {
  apiService.loginWithGithub();
  console.log("called once");
};
</script>
