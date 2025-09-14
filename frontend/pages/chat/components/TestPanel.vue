<template>
  <div class="h-full flex flex-col">
    <div class="mb-4 md:mb-6">
      <h2 class="text-lg md:text-xl font-semibold mb-1">Test Your Chatbot</h2>
      <p class="text-sm text-muted-foreground">
        <span class="hidden md:inline">
          Make changes on the left and test them here in real-time
        </span>
        <span class="md:hidden">Test your chatbot configuration here</span>
      </p>
    </div>

    <!-- Loading skeleton -->
    <div v-if="isLoadingChatbot" class="animate-pulse space-y-4">
      <div class="h-5 w-32 md:w-40 bg-muted rounded"></div>
      <div class="h-24 md:h-28 bg-muted/70 rounded"></div>
      <div class="h-10 bg-muted rounded"></div>
    </div>

    <!-- Knowledge required state -->
    <div
      v-else-if="!hasKnowledgeBaseData"
      class="mt-7 flex items-center justify-center"
    >
      <div class="text-center py-6 md:py-8">
        <div
          class="mx-auto inline-flex h-10 md:h-12 w-10 md:w-12 items-center justify-center rounded-xl md:rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-500 text-white shadow-sm"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            class="h-5 md:h-6 w-5 md:w-6"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" />
            <path d="M4 4v16" />
          </svg>
        </div>
        <h3 class="mt-3 md:mt-4 text-base md:text-lg font-medium">
          Add Knowledge to Start<span class="hidden md:inline"> Chatting</span>
        </h3>
        <p class="mt-1 md:mt-2 text-sm text-muted-foreground">
          Add
          <span class="hidden md:inline"
            >files, text, or websites to your knowledge base</span
          >
          <span class="md:hidden">data sources</span> before starting a
          conversation.
        </p>
        <div class="mt-4 md:mt-5">
          <Button
            :size="isMobile ? 'sm' : 'default'"
            variant="secondary"
            @click="scrollToKnowledge"
          >
            Add Data Sources
          </Button>
        </div>
      </div>
    </div>

    <!-- Chat Interface -->
    <ChatInterface
      v-else
      :class="isMobile ? 'chat-interface-mobile' : 'chat-interface'"
      :chatbot="chatbot"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { useBreakpoints } from "@vueuse/core";
import { Button } from "@/components/ui/button";
import ChatInterface from "./ChatInterface.vue";
import type { ChatbotResponse } from "~/types/api";
import { useGlobalState } from "@/composables/useGlobalState";

// Route & API
const route = useRoute();
const router = useRouter();
const apiService = useApiService();
const chatId = computed(() => route.params.id as string);
const { hasKnowledgeBaseData } = useGlobalState();

// State
const chatbot = ref<ChatbotResponse | null>(null);

// API calls
const {
  data,
  execute: executeFetchChatbot,
  error: fetchChatbotError,
  isLoading: isLoadingChatbot,
} = apiService.getChatbot();

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

// Scroll to knowledge section in details page
const scrollToKnowledge = () => {
  router.push(`/chat/${chatId.value}/details`);
};

const breakpoints = useBreakpoints({ md: 768 });
const isMobile = computed(() => !breakpoints.greaterOrEqual("md").value);

const chatInterface = ref<InstanceType<typeof ChatInterface> | null>(null);

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
});

defineExpose({
  chatInterface,
});
</script>

<style scoped>
.chat-interface-mobile {
  height: calc(100vh - 300px);
  max-height: 500px;
}

.chat-interface {
  height: calc(100vh - 250px);
  max-height: 700px;
}
</style>
