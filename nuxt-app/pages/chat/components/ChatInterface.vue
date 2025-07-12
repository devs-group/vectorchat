<template>
  <div class="flex flex-col w-full max-w-xl justify-start">
    <!-- Header with Chat Information -->
    <h1 class="mb-6 text-2xl font-bold tracking-tight text-left">
      {{ chatbot?.name || "Chat" }}
    </h1>
    <p v-if="chatbot" class="text-muted-foreground text-sm mb-4 text-left">
      {{ chatbot.description }}
    </p>

    <!-- Messages Container -->
    <div
      class="flex-1 overflow-y-auto rounded border p-2 bg-white/70 min-h-[200px] max-h-[400px] mb-4"
    >
      <div
        v-if="messages.length === 0"
        class="flex flex-col items-center justify-center h-full py-8"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="32"
          height="32"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1"
          class="mb-2 text-muted-foreground"
        >
          <path
            d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"
          ></path>
        </svg>
        <h3 class="font-medium text-base mb-0.5">No messages yet</h3>
        <p class="text-xs text-muted-foreground mb-2">
          Start a conversation with your AI assistant
        </p>
      </div>
      <div v-else class="flex flex-col gap-2">
        <div
          v-for="(message, index) in messages"
          :key="index"
          :class="[
            'rounded p-2',
            message.isUser ? 'bg-muted ml-auto' : 'bg-primary/10 mr-auto',
          ]"
          style="max-width: 75%"
        >
          <div class="flex items-center gap-1 mb-0.5">
            <div
              class="h-6 w-6 rounded-full flex items-center justify-center"
              :class="message.isUser ? 'bg-secondary' : 'bg-primary'"
            >
              <svg
                v-if="message.isUser"
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                class="h-3 w-3 text-background"
              >
                <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"></path>
                <circle cx="12" cy="7" r="4"></circle>
              </svg>
              <svg
                v-else
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                class="h-3 w-3 text-background"
              >
                <path d="M12 8V4H8"></path>
                <rect
                  x="2"
                  y="2"
                  width="20"
                  height="20"
                  rx="2.18"
                  ry="2.18"
                ></rect>
                <path d="M10.14 15.25a3 3 0 0 0 4.3-1.2"></path>
              </svg>
            </div>
            <span class="font-medium text-xs">{{
              message.isUser ? "You" : chatbot?.name || "AI"
            }}</span>
            <span class="text-[10px] text-muted-foreground">{{
              message.timestamp
            }}</span>
          </div>
          <div class="text-sm whitespace-pre-wrap text-left">
            {{ message.content }}
          </div>
        </div>
      </div>
    </div>

    <!-- Input Box -->
    <div class="mt-2">
      <div class="relative">
        <textarea
          v-model="newMessage"
          class="w-full rounded border px-3 py-2 pr-12 resize-none text-sm min-h-[36px] max-h-[80px]"
          rows="1"
          placeholder="Type your message..."
          @keydown.enter.prevent="sendMessage"
        ></textarea>
        <button
          class="absolute right-2 top-2 text-primary hover:text-primary/70 transition-colors"
          @click="sendMessage"
          :disabled="isSendingMessage || !newMessage.trim()"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="h-5 w-5"
            :class="{ 'opacity-50': isSendingMessage || !newMessage.trim() }"
          >
            <path d="M12 19V5"></path>
            <path d="m5 12 7-7 7 7"></path>
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { toast } from "vue-sonner";
import { Button } from "@/components/ui/button";
import type { ChatbotResponse } from "~/types/api";

interface Props {
  chatId: string;
}

interface Message {
  content: string;
  isUser: boolean;
  timestamp: string;
}

const props = defineProps<Props>();

// API service
const apiService = useApiService();

// Chat data
const chatbot = ref<ChatbotResponse | null>(null);
const messages = ref<Message[]>([]);
const newMessage = ref("");

// Loading state
const isSendingMessage = ref(false);

// Fetch chatbot details
const {
  data: chatbotData,
  execute: fetchChatbot,
  error: chatbotError,
  isLoading: isLoadingChatbot,
} = apiService.getChatbot(props.chatId);

const fetchChatbotDetails = async () => {
  try {
    await fetchChatbot();

    if (chatbotData.value?.chatbot) {
      chatbot.value = chatbotData.value.chatbot;
    } else {
      console.error("Chat not found");
      throw new Error("Chat not found");
    }
  } catch (error) {
    console.error("Error fetching chatbot details:", error);
    throw error;
  }
};

// Send a message
const sendMessage = async () => {
  if (!newMessage.value.trim() || isSendingMessage.value) return;

  const userMessage = newMessage.value.trim();

  // Add user message to the messages array
  messages.value.push({
    content: userMessage,
    isUser: true,
    timestamp: new Date().toLocaleTimeString(),
  });

  // Clear input
  newMessage.value = "";

  // Send message to API
  isSendingMessage.value = true;

  try {
    const { data: responseData, execute: executeSendMessage } =
      apiService.sendChatMessage(props.chatId, userMessage);

    await executeSendMessage();

    if (responseData.value && typeof responseData.value === "object") {
      // Add AI response to messages array
      const responseMessage =
        "message" in responseData.value
          ? responseData.value.message
          : "response" in responseData.value
            ? responseData.value.response
            : "I processed your message, but I have no specific response.";

      messages.value.push({
        content: responseMessage as string,
        isUser: false,
        timestamp: new Date().toLocaleTimeString(),
      });
    }
  } catch (error) {
    console.error("Error sending message:", error);

    // Add an error message
    messages.value.push({
      content:
        "Sorry, there was an error processing your message. Please try again.",
      isUser: false,
      timestamp: new Date().toLocaleTimeString(),
    });
  } finally {
    isSendingMessage.value = false;
  }
};

// Initialize chat data
const initializeChat = async () => {
  try {
    await fetchChatbotDetails();
  } catch (error) {
    console.error("Error initializing chat:", error);
    throw error;
  }
};

// Reset chat data
const resetChat = () => {
  messages.value = [];
  chatbot.value = null;
  newMessage.value = "";
};

// Watch for chatId changes
watch(
  () => props.chatId,
  async (newChatId, oldChatId) => {
    if (newChatId && newChatId !== oldChatId) {
      resetChat();
      await initializeChat();
    }
  },
);

// Initialize on mount
onMounted(async () => {
  await initializeChat();
});

// Expose methods for parent component
defineExpose({
  initializeChat,
  resetChat,
  chatbot,
  messages,
});
</script>
