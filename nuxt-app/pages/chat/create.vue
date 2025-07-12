<script setup lang="ts">
definePageMeta({
  layout: "authenticated",
});
import ChatbotForm from "./components/ChatbotForm.vue";

const router = useRouter();
const apiService = useApiService();

const isLoading = ref(false);

const handleSubmit = async (formData: any) => {
  isLoading.value = true;
  try {
    const { execute, data, error } = apiService.createChatbot(formData);
    await execute();
    if (error.value) throw error.value;
    toast.success("Chatbot created successfully!");
    router.push("/chat");
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
