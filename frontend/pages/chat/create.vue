<script setup lang="ts">
definePageMeta({
  layout: "authenticated",
});
import { computed, onMounted } from "vue";
import BackButton from "@/components/BackButton.vue";
import ChatbotForm from "./components/ChatbotForm.vue";
import type {
  ChatbotResponse,
  SharedKnowledgeBaseListResponse,
} from "~/types/api";

const router = useRouter();
const apiService = useApiService();
const { execute, data, error, isLoading } = apiService.createChatbot();
const { execute: loadSharedKnowledgeBases, data: sharedKnowledgeBasesData } =
  apiService.listSharedKnowledgeBases();

onMounted(async () => {
  try {
    await loadSharedKnowledgeBases();
  } catch (err) {
    console.error("Failed to load shared knowledge bases", err);
  }
});

const availableSharedKnowledgeBases = computed(() => {
  const response = sharedKnowledgeBasesData.value as
    | SharedKnowledgeBaseListResponse
    | undefined;
  return response?.knowledge_bases ?? [];
});

const handleSubmit = async (formData: any) => {
  await execute(formData);
  if (!error.value) {
    router.push({ path: `/chat/${(data.value as ChatbotResponse).id}` });
  }
};
</script>

<template>
  <div class="flex min-h-screen bg-background px-4">
    <div class="flex w-full flex-col">
      <BackButton fallback="/chat" class="mb-6" />
      <div class="flex flex-1">
        <ChatbotForm
          mode="create"
          :is-loading="isLoading"
          :shared-knowledge-bases="availableSharedKnowledgeBases"
          @submit="handleSubmit"
        />
        <!-- Empty right side for future content -->
        <div class="flex-1"></div>
      </div>
    </div>
  </div>
</template>
