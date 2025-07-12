<template>
  <div class="flex min-h-screen bg-background">
    <!-- Left Side - Edit Form -->
    <div class="w-1/2 border-r border-border bg-background px-6 py-8">
      <div class="max-w-lg mx-auto">
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
          <Button @click="fetchChatbotData" variant="outline">Try Again</Button>
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
        <div v-if="chatbot" class="mt-8 pt-8 border-t border-border">
          <FileUpload :chat-id="chatId" ref="fileUpload" />
        </div>
      </div>
    </div>

    <!-- Right Side - Chat Interface -->
    <div class="w-1/2 bg-muted/20 px-6 py-8">
      <div class="max-w-lg mx-auto">
        <div class="mb-6">
          <h2 class="text-xl font-semibold mb-2">Test Your Chatbot</h2>
          <p class="text-sm text-muted-foreground">
            Make changes on the left and test them here in real-time
          </p>
        </div>

        <!-- Chat Interface -->
        <ChatInterface
          v-if="chatId"
          :chat-id="chatId"
          @error="handleChatError"
          ref="chatInterface"
        />

        <!-- API Documentation -->
        <ApiDocumentation v-if="chatId" :chat-id="chatId" />
      </div>
    </div>

    <!-- Update Success Toast -->
    <div
      v-if="showUpdateSuccess"
      class="fixed top-4 right-4 bg-green-500 text-white px-4 py-2 rounded-md shadow-lg z-50 transition-all duration-300"
    >
      Chatbot updated successfully!
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from "vue";
import { toast } from "vue-sonner";
import ChatInterface from "./components/ChatInterface.vue";
import ChatbotForm from "./components/ChatbotForm.vue";
import FileUpload from "./components/FileUpload.vue";
import ApiDocumentation from "./components/ApiDocumentation.vue";
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

// Refs
const chatInterface = ref<InstanceType<typeof ChatInterface> | null>(null);
const fileUpload = ref<InstanceType<typeof FileUpload> | null>(null);

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
onMounted(() => {
  fetchChatbotData();
});
</script>

<style scoped>
/* Custom scrollbar for better aesthetics */
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
