<template>
  <div class="max-w-3xl mx-auto">
    <!-- Loading State -->
    <div v-if="isLoadingChatbot" class="flex items-center justify-center py-12">
      <div
        class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
      ></div>
    </div>

    <!-- Edit Form -->
    <div v-else-if="chatbot">
      <!-- Enable/Disable Toggle -->
      <div class="mb-6 p-4 rounded-lg border border-border bg-card">
        <div class="flex items-center justify-between">
          <div class="space-y-0.5">
            <h3 class="font-medium">Enabled</h3>
            <p class="text-sm text-muted-foreground">
              Toggle chatbot for being enabled or disabled
            </p>
          </div>
          <Switch
            :model-value="chatbot.is_enabled"
            @update:model-value="handleToggleEnabled"
          />
        </div>
      </div>

      <ChatbotForm
        mode="edit"
        :chatbot="chatbot"
        :is-loading="isUpdating"
        :shared-knowledge-bases="sharedKnowledgeBases"
        @submit="handleUpdate"
      />
    </div>

    <!-- File Upload Section -->
    <div
      v-if="chatbot"
      id="knowledgeSection"
      class="mt-8 pt-8 border-t border-border"
    >
      <KnowledgeBase :resource-id="chatId" scope="chatbot" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from "vue";
import ChatbotForm from "../components/ChatbotForm.vue";
import KnowledgeBase from "./components/KnowledgeBase.vue";
import { Switch } from "@/components/ui/switch";
import type {
  ChatbotResponse,
  SharedKnowledgeBaseListResponse,
} from "~/types/api";
import { useRoute } from "vue-router";
import { useApiService } from "@/composables/useApiService";

// Route & API
const route = useRoute();
const apiService = useApiService();
const chatId = computed(() => route.params.id as string);

// State
const chatbot = ref<ChatbotResponse | null>(null);
const isToggling = ref(false);

// Refs
const knowledgeSection = ref<HTMLElement | null>(null);

// API calls
const {
  data,
  execute: executeFetchChatbot,
  error: fetchChatbotError,
  isLoading: isLoadingChatbot,
} = apiService.getChatbot();

const { execute: executeToggle, error: errorToggle } =
  apiService.toggleChatbot();

const {
  execute: executeUpdate,
  error: updateError,
  isLoading: isUpdating,
} = apiService.updateChatbot();

const { execute: loadSharedKnowledgeBases, data: sharedKnowledgeBasesData } =
  apiService.listSharedKnowledgeBases();

const sharedKnowledgeBases = computed(() => {
  const response = sharedKnowledgeBasesData.value as
    | SharedKnowledgeBaseListResponse
    | undefined;
  return response?.knowledge_bases ?? [];
});

// Fetch chatbot data
const fetchChatbotData = async () => {
  if (!chatId.value) return;

  await executeFetchChatbot(chatId.value);
  if (fetchChatbotError.value) {
    return;
  }
  if (data.value?.chatbot) {
    chatbot.value = data.value.chatbot;
  }
};

// Handle toggle enabled/disabled
const handleToggleEnabled = async () => {
  if (!chatId.value || !chatbot.value) return;

  isToggling.value = true;
  const newEnabledState = !chatbot.value.is_enabled;

  await executeToggle({
    chatbotId: chatbot.value.id,
    isEnabled: newEnabledState,
  });

  isToggling.value = false;

  if (!errorToggle.value) {
    chatbot.value.is_enabled = newEnabledState;
  }
};

// Handle update
const handleUpdate = async (formData: any) => {
  if (!chatId.value) return;

  await executeUpdate({
    id: chatId.value,
    ...formData,
  });

  if (!updateError.value && chatbot.value) {
    chatbot.value = { ...chatbot.value, ...formData };
  }
};

// Scroll to knowledge section
const scrollToKnowledge = async () => {
  await nextTick();
  if (knowledgeSection.value) {
    knowledgeSection.value.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  }
};

// Watch for route changes
watch(
  () => route.params.id,
  (newId) => {
    if (newId) {
      fetchChatbotData();
    }
  },
);

// Initialize
onMounted(() => {
  fetchChatbotData();
  loadSharedKnowledgeBases();
});

// Expose scroll function for external use
defineExpose({
  scrollToKnowledge,
  knowledgeSection,
});
</script>
