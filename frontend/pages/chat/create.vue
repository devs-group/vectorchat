<script setup lang="ts">
definePageMeta({
  layout: "authenticated",
});
import ChatbotForm from "./components/ChatbotForm.vue";
import type { ChatbotResponse } from "~/types/api";

const router = useRouter();
const apiService = useApiService();
const { execute, data, error, isLoading } = apiService.createChatbot();

const handleSubmit = async (formData: any) => {
  await execute(formData);
  if (!error.value) {
    router.push({ path: `/chat/${(data.value as ChatbotResponse).id}` });
  }
};
</script>

<template>
  <div class="flex min-h-screen bg-background px-4 py-12">
    <ChatbotForm mode="create" :is-loading="isLoading" @submit="handleSubmit" />
    <!-- Empty right side for future content -->
    <div class="flex-1"></div>
  </div>
</template>
