<template>
  <div class="min-h-screen bg-background">
    <!-- Mobile Tab Navigation -->
    <div class="md:hidden bg-background border-b border-border">
      <div class="flex">
        <button
          @click="activeTab = 'edit'"
          :class="[
            'flex-1 py-4 px-6 text-sm font-medium text-center border-b-2 transition-colors',
            activeTab === 'edit'
              ? 'border-primary text-primary bg-primary/5'
              : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border',
          ]"
        >
          Edit Chatbot
        </button>
        <button
          @click="activeTab = 'test'"
          :class="[
            'flex-1 py-4 px-6 text-sm font-medium text-center border-b-2 transition-colors',
            activeTab === 'test'
              ? 'border-primary text-primary bg-primary/5'
              : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border',
          ]"
        >
          Test Chatbot
        </button>
      </div>
    </div>

    <!-- Desktop Layout & Mobile Content -->
    <div class="md:flex min-h-screen">
      <!-- Left Side - Edit Form -->
      <div
        :class="[
          'md:w-1/2 md:border-r border-border bg-background px-4 md:px-6 py-6 md:py-8',
          activeTab === 'edit' || 'md:block hidden',
        ]"
      >
        <div class="max-w-3xl mx-auto">
          <!-- Loading State -->
          <div
            v-if="isLoadingChatbot"
            class="flex items-center justify-center py-12"
          >
            <div
              class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
            ></div>
          </div>

          <!-- Error State -->
          <div v-else-if="chatbotError" class="text-center py-12">
            <div class="text-destructive mb-4">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="48"
                height="48"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                class="mx-auto mb-2"
              >
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="15" y1="9" x2="9" y2="15"></line>
                <line x1="9" y1="9" x2="15" y2="15"></line>
              </svg>
            </div>
            <h3 class="text-lg font-medium mb-2">Error Loading Chatbot</h3>
            <p class="text-muted-foreground mb-4">{{ chatbotError.message }}</p>
            <Button @click="fetchChatbotData" variant="outline"
              >Try Again</Button
            >
          </div>

          <!-- Edit Form -->
          <ChatbotForm
            v-else-if="chatbot"
            mode="edit"
            :chatbot="chatbot"
            :is-loading="isUpdating"
            @submit="handleUpdate"
          />

          <!-- File Upload Section -->
          <div
            v-if="chatbot"
            ref="knowledgeSection"
            class="mt-8 pt-8 border-t border-border"
          >
            <FileUpload :chat-id="chatId" ref="fileUpload" />
          </div>
        </div>
      </div>

      <!-- Right Side - Chat Interface -->
      <div
        :class="[
          'md:w-1/2 bg-muted/20 px-4 md:px-6 py-6 md:py-8',
          activeTab === 'test' || 'md:block hidden',
        ]"
      >
        <div class="max-w-3xl mx-auto">
          <div class="mb-6">
            <h2 class="text-xl font-semibold mb-1">Test Your Chatbot</h2>
            <p class="text-sm text-muted-foreground">
              <span class="hidden md:inline"
                >Make changes on the left and test them here in real-time</span
              >
              <span class="md:hidden"
                >Test your chatbot configuration here</span
              >
            </p>
          </div>

          <!-- Loading skeleton for test panel -->
          <div
            v-if="isLoadingChatbot"
            class="rounded-2xl border border-border bg-card p-6"
          >
            <div class="animate-pulse space-y-4">
              <div class="h-5 w-40 bg-muted rounded"></div>
              <div class="h-28 bg-muted/70 rounded"></div>
              <div class="h-10 bg-muted rounded"></div>
            </div>
          </div>

          <!-- Knowledge required state -->
          <div
            v-else-if="chatbot && knowledgeChecked && !hasKnowledgeBase"
            class="rounded-2xl border border-border bg-card p-6 md:p-8 text-center shadow-sm"
          >
            <div
              class="mx-auto inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-500 text-white shadow-sm"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                class="h-6 w-6"
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
            <h3 class="mt-4 text-lg font-medium">
              Add Knowledge to Start Chatting
            </h3>
            <p class="mt-2 text-sm text-muted-foreground">
              Add files, text, or websites to your knowledge base before
              starting a conversation.
            </p>
            <div class="mt-5 flex items-center justify-center gap-3">
              <Button variant="secondary" @click="scrollToKnowledge"
                >Add Data Sources</Button
              >
            </div>
          </div>

          <!-- Chat Interface -->
          <div
            v-else-if="chatId && hasKnowledgeBase"
            class="rounded-2xl border border-border bg-card shadow-sm p-6 md:p-8"
          >
            <ChatInterface
              class="chat-interface"
              :chat-id="chatId"
              @error="handleChatError"
              ref="chatInterface"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- Update Success Toast -->
    <div
      v-if="showUpdateSuccess"
      class="fixed top-4 right-4 md:top-4 md:right-4 left-4 md:left-auto bg-green-500 text-white px-4 py-2 rounded-md shadow-lg z-50 transition-all duration-300"
    >
      Chatbot updated successfully!
    </div>

    <!-- Mobile Floating Action Button for Quick Switch -->
    <div class="md:hidden fixed bottom-6 right-6 z-40">
      <button
        @click="toggleTab"
        class="bg-primary hover:bg-primary/90 text-primary-foreground w-14 h-14 rounded-full shadow-lg flex items-center justify-center transition-all duration-200 active:scale-95"
      >
        <svg
          v-if="activeTab === 'edit'"
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path
            d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"
          ></path>
        </svg>
        <svg
          v-else
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M12 20h9"></path>
          <path
            d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"
          ></path>
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from "vue";
import { toast } from "vue-sonner";
import ChatInterface from "./components/ChatInterface.vue";
import ChatbotForm from "./components/ChatbotForm.vue";
import FileUpload from "./components/FileUpload.vue";
import { Button } from "@/components/ui/button";
import type { ChatbotResponse } from "~/types/api";

definePageMeta({
  layout: "authenticated",
});

// Get the chat ID from the route
const route = useRoute();
const router = useRouter();
const chatId = ref(route.params.id as string);

// API service
const apiService = useApiService();

// State
const chatbot = ref<ChatbotResponse | null>(null);
const isLoadingChatbot = ref(false);
const isUpdating = ref(false);
const chatbotError = ref<Error | null>(null);
const showUpdateSuccess = ref(false);
const activeTab = ref<"edit" | "test">("edit");

// Refs
const chatInterface = ref<InstanceType<typeof ChatInterface> | null>(null);
const fileUpload = ref<InstanceType<typeof FileUpload> | null>(null);
const knowledgeSection = ref<HTMLElement | null>(null);
const hasKnowledgeBase = ref(false);
const knowledgeChecked = ref(false);

const updateKnowledgeFromChild = () => {
  const filesLen = Array.isArray(fileUpload.value?.files)
    ? (fileUpload.value!.files as any as any[]).length
    : 0;
  const textLen = Array.isArray((fileUpload.value as any)?.textSources)
    ? ((fileUpload.value as any).textSources as any[]).length
    : 0;
  hasKnowledgeBase.value = filesLen + textLen > 0;
  knowledgeChecked.value = true;
};

// Fetch chatbot data
const fetchChatbotData = async () => {
  if (!chatId.value) return;

  isLoadingChatbot.value = true;
  chatbotError.value = null;

  try {
    const { data, execute, error } = apiService.getChatbot(chatId.value);
    await execute();

    if (error.value) {
      throw error.value;
    }

    if (data.value?.chatbot) {
      chatbot.value = data.value.chatbot;
    } else {
      throw new Error("Chatbot not found");
    }
  } catch (err) {
    console.error("Error fetching chatbot:", err);
    chatbotError.value = err as Error;

    if ((err as Error).message.includes("not found")) {
      // Redirect to chat list if chatbot doesn't exist
      setTimeout(() => {
        router.push("/chat");
      }, 2000);
    }
  } finally {
    isLoadingChatbot.value = false;
  }
};

// Handle chatbot update
const handleUpdate = async (formData: any) => {
  if (!chatId.value) return;

  isUpdating.value = true;

  try {
    const { execute, error } = apiService.updateChatbot(chatId.value, formData);
    await execute();

    if (error.value) {
      throw error.value;
    }

    // Update local chatbot data
    if (chatbot.value) {
      chatbot.value = { ...chatbot.value, ...formData };
    }

    // Show success message
    showUpdateSuccess.value = true;
    setTimeout(() => {
      showUpdateSuccess.value = false;
    }, 3000);

    // Refresh chat interface to reflect changes
    if (chatInterface.value) {
      await chatInterface.value.initializeChat();
    }

    toast.success("Chatbot updated successfully!");
  } catch (err: any) {
    console.error("Error updating chatbot:", err);
    toast.error("Failed to update chatbot", {
      description: err?.message || "An error occurred",
    });
  } finally {
    isUpdating.value = false;
  }
};

// Handle chat errors
const handleChatError = (error: Error) => {
  console.error("Chat error:", error);
  if (error.message === "Chat not found") {
    router.push("/chat");
  }
};

// Toggle between tabs on mobile
const toggleTab = () => {
  activeTab.value = activeTab.value === "edit" ? "test" : "edit";
};

// Smoothly jump user to the knowledge base section
const scrollToKnowledge = async () => {
  if (activeTab.value !== "edit") {
    activeTab.value = "edit";
  }
  await nextTick();
  knowledgeSection.value?.scrollIntoView({
    behavior: "smooth",
    block: "start",
  });
};

// Watch for route changes
watch(
  () => route.params.id,
  (newId) => {
    if (newId) {
      chatId.value = newId as string;
      fetchChatbotData();
    }
  },
);

// Initialize on mount
onMounted(async () => {
  fetchChatbotData();
  await nextTick();
});

// Watch child sources to keep knowledge state in sync
watch(
  () => [
    fileUpload.value?.files.length,
    (fileUpload.value as any)?.textSources?.length,
  ],
  () => updateKnowledgeFromChild(),
  { immediate: true },
);

// Also update when the child ref attaches
watch(
  () => fileUpload.value,
  (val) => {
    if (val) updateKnowledgeFromChild();
  },
  { immediate: true },
);
</script>

<style scoped>
/* Custom scrollbar for better aesthetics */
:deep(.overflow-y-auto) {
  scrollbar-width: thin;
  scrollbar-color: rgba(156, 163, 175, 0.5) transparent;
}

/* Prevent horizontal scroll on mobile */
@media (max-width: 768px) {
  body {
    overflow-x: hidden;
  }
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

/* Mobile-specific styles */
@media (max-width: 768px) {
  /* Ensure full height on mobile */
  .min-h-screen {
    min-height: 100vh;
    min-height: 100dvh; /* Dynamic viewport height for mobile browsers */
  }

  /* Better spacing on mobile */
  .max-w-lg {
    max-width: 100%;
    padding: 0 1rem;
  }

  /* Adjust padding for mobile */
  .px-4 {
    padding-left: 1rem;
    padding-right: 1rem;
  }

  /* Improve form spacing on mobile */
  :deep(.space-y-4 > *) {
    margin-bottom: 1rem;
  }

  /* Better button sizing on mobile */
  :deep(button) {
    min-height: 44px; /* Minimum touch target size */
    padding: 0.75rem 1rem;
  }

  /* Improve input sizing on mobile */
  :deep(input),
  :deep(textarea),
  :deep(select) {
    min-height: 44px;
    font-size: 16px; /* Prevent zoom on iOS */
  }

  /* Improve chat interface on mobile */
  :deep(.chat-interface) {
    height: calc(100vh - 180px);
    height: calc(100dvh - 180px);
    max-height: 70vh;
  }

  /* Better tab styling on mobile */
  .border-b-2 {
    border-bottom-width: 3px;
  }

  /* Improve form layout on mobile */
  :deep(.grid) {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
}

/* Smooth transitions for tab switching */
.tab-content {
  transition: opacity 0.2s ease-in-out;
}

/* Improve touch targets on mobile */
@media (max-width: 768px) {
  button,
  .clickable {
    min-height: 44px;
    min-width: 44px;
  }
}
</style>
