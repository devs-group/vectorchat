<script setup lang="ts">
definePageMeta({
  layout: "authenticated",
});
import ChatbotForm from "./components/ChatbotForm.vue";
import { toast } from "vue-sonner";
import type { ChatbotResponse } from "~/types/api";

const router = useRouter();
const apiService = useApiService();

const isLoading = ref(false);

const handleSubmit = async (formData: any) => {
  isLoading.value = true;
  try {
    const { execute, data, error } = apiService.createChatbot(formData);
    await execute();
    if (error.value) throw error.value;

    // Get the created chatbot ID from the response
    const chatbotId = (data.value as ChatbotResponse)?.id;

    if (!chatbotId) {
      console.error("No chatbot ID returned from API response:", data.value);
      toast.error("Created chatbot but failed to get ID for redirection");
      router.push("/chat");
      return;
    }

    // Redirect to the detail page of the newly created chatbot
    console.log("Redirecting to chatbot detail page:", chatbotId);
    router.push(`/chat/${chatbotId}`);
  } catch (err: any) {
    toast.error("Failed to create chatbot", {
      description: err?.message || "An error occurred",
    });
  } finally {
    isLoading.value = false;
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
