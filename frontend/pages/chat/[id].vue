<template>
  <div class="min-h-screen bg-background">
    <div class="container mx-auto p-4 md:p-6">
      <!-- Desktop: Two column layout -->
      <div
        class="hidden md:grid md:grid-cols-2 md:gap-6 min-h-[calc(100vh-3rem)]"
      >
        <!-- Left column with tabs -->
        <div class="border rounded-lg bg-card overflow-hidden">
          <Tabs v-model="activeTab" class="h-full flex flex-col">
            <TabsList class="grid w-full grid-cols-2 rounded-none">
              <TabsTrigger value="details">Details</TabsTrigger>
              <TabsTrigger value="history">Chat History</TabsTrigger>
            </TabsList>

            <TabsContent
              value="details"
              class="flex-1 overflow-y-auto p-6 mt-0"
            >
              <ChatbotDetails
                :chatbot="chatbot"
                :chat-id="chatId"
                :is-loading-chatbot="isLoadingChatbot"
                :chatbot-error="chatbotError"
                :is-toggling="isToggling"
                :is-updating="isUpdating"
                @toggle-enabled="handleToggleEnabled"
                @update="handleUpdate"
                @retry="fetchChatbotData"
                @knowledge-update="handleKnowledgeUpdate"
                ref="chatbotDetailsRef"
              />
            </TabsContent>

            <TabsContent
              value="history"
              class="flex-1 overflow-y-auto p-6 mt-0"
            >
              <ChatHistory :chat-id="chatId" @switch-to-test="() => {}" />
            </TabsContent>
          </Tabs>
        </div>

        <!-- Right column with test panel -->
        <div class="border rounded-lg bg-card p-6 overflow-hidden">
          <TestPanel
            :chatbot="chatbot"
            :chat-id="chatId"
            :is-loading-chatbot="isLoadingChatbot"
            :has-knowledge-base="hasKnowledgeBase"
            :knowledge-checked="knowledgeChecked"
            @scroll-to-knowledge="scrollToKnowledge"
            @chat-error="handleChatError"
          />
        </div>
      </div>

      <!-- Mobile: Single column with tabs -->
      <div class="md:hidden">
        <Tabs v-model="activeTab" class="w-full">
          <TabsList class="grid w-full grid-cols-3">
            <TabsTrigger value="details">Details</TabsTrigger>
            <TabsTrigger value="history">History</TabsTrigger>
            <TabsTrigger value="test">Test</TabsTrigger>
          </TabsList>

          <TabsContent value="details" class="mt-4">
            <ChatbotDetails
              :chatbot="chatbot"
              :chat-id="chatId"
              :is-loading-chatbot="isLoadingChatbot"
              :chatbot-error="chatbotError"
              :is-toggling="isToggling"
              :is-updating="isUpdating"
              @toggle-enabled="handleToggleEnabled"
              @update="handleUpdate"
              @retry="fetchChatbotData"
              @knowledge-update="handleKnowledgeUpdate"
              ref="chatbotDetailsRef"
            />
          </TabsContent>

          <TabsContent value="history" class="mt-4">
            <ChatHistory
              :chat-id="chatId"
              @switch-to-test="() => (activeTab = 'test')"
            />
          </TabsContent>

          <TabsContent value="test" class="mt-4">
            <TestPanel
              :chatbot="chatbot"
              :chat-id="chatId"
              :is-loading-chatbot="isLoadingChatbot"
              :has-knowledge-base="hasKnowledgeBase"
              :knowledge-checked="knowledgeChecked"
              @scroll-to-knowledge="scrollToKnowledge"
              @chat-error="handleChatError"
            />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from "vue";
import { useBreakpoints } from "@vueuse/core";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import ChatbotDetails from "./components/ChatbotDetails.vue";
import ChatHistory from "./components/ChatHistory.vue";
import TestPanel from "./components/TestPanel.vue";
import type { ChatbotResponse } from "~/types/api";

definePageMeta({
  layout: "authenticated",
});

// Route & API
const route = useRoute();
const router = useRouter();
const apiService = useApiService();
const chatId = computed(() => route.params.id as string);

// Responsive
const breakpoints = useBreakpoints({ md: 768 });
const isMobile = computed(() => !breakpoints.greaterOrEqual("md").value);

// State
const chatbot = ref<ChatbotResponse | null>(null);
const chatbotError = ref<Error | null>(null);
const isToggling = ref(false);
const hasKnowledgeBase = ref(false);
const knowledgeChecked = ref(false);
const activeTab = ref(isMobile.value ? "details" : "details");

// Refs
const chatbotDetailsRef = ref<InstanceType<typeof ChatbotDetails> | null>(null);

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

// Fetch chatbot data
const fetchChatbotData = async () => {
  if (!chatId.value) return;

  await executeFetchChatbot(chatId.value);
  if (fetchChatbotError.value) {
    chatbotError.value = fetchChatbotError.value;
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

// Handle knowledge update
const handleKnowledgeUpdate = (hasKnowledge: boolean) => {
  hasKnowledgeBase.value = hasKnowledge;
  knowledgeChecked.value = true;
};

// Handle chat error
const handleChatError = (error: Error) => {
  console.error("Chat error:", error);
  if (error.message === "Chat not found") {
    router.push("/chat");
  }
};

// Scroll to knowledge section
const scrollToKnowledge = async () => {
  if (isMobile.value) {
    activeTab.value = "details";
  } else {
    activeTab.value = "details";
  }

  await nextTick();

  const knowledgeSection = chatbotDetailsRef.value?.knowledgeSection;
  knowledgeSection?.scrollIntoView({
    behavior: "smooth",
    block: "start",
  });
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

// Watch for knowledge base changes from child
watch(
  () => chatbotDetailsRef.value?.fileUpload,
  (fileUpload) => {
    if (!fileUpload) return;

    watch(
      () => [fileUpload.files?.length, (fileUpload as any).textSources?.length],
      () => {
        const filesLen = Array.isArray(fileUpload.files)
          ? fileUpload.files.length
          : 0;
        const textLen = Array.isArray((fileUpload as any)?.textSources)
          ? (fileUpload as any).textSources.length
          : 0;
        hasKnowledgeBase.value = filesLen + textLen > 0;
        knowledgeChecked.value = true;
      },
      { immediate: true },
    );
  },
  { immediate: true },
);

// Initialize
onMounted(() => {
  fetchChatbotData();
});
</script>

<style scoped>
/* Custom scrollbar */
:deep(.overflow-y-auto) {
  scrollbar-width: thin;
  scrollbar-color: rgba(156, 163, 175, 0.5) transparent;
}

:deep(.overflow-y-auto::-webkit-scrollbar) {
  width: 6px;
}

:deep(.overflow-y-auto::-webkit-scrollbar-track) {
  background: transparent;
}

:deep(.overflow-y-auto::-webkit-scrollbar-thumb) {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 3px;
}

:deep(.overflow-y-auto::-webkit-scrollbar-thumb:hover) {
  background-color: rgba(156, 163, 175, 0.7);
}
</style>
