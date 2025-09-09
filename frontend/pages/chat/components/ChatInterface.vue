<template>
  <div class="flex flex-col w-full max-w-xl justify-start">
    <!-- Header -->
    <h1 class="mb-2 text-xl font-semibold tracking-tight text-left">
      {{ chatbot?.name || "Chat" }}
    </h1>
    <p v-if="chatbot" class="text-muted-foreground text-sm mb-4 text-left">
      {{ chatbot.description }}
    </p>

    <!-- Messages Container -->
    <div
      ref="messagesContainer"
      class="flex-1 overflow-y-auto rounded-xl border border-border bg-muted/20 p-3 md:p-4 min-h-[220px] max-h-[420px] mb-4"
    >
      <div
        v-if="messages.length === 0 && !isSendingMessage"
        class="flex flex-col items-center justify-center h-full py-8 text-center"
      >
        <div class="mx-auto inline-flex h-10 w-10 items-center justify-center rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-500 text-white shadow-sm">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        </div>
        <h3 class="mt-3 font-medium text-base">No messages yet</h3>
        <p class="text-xs text-muted-foreground mt-1">Start a conversation with your AI assistant</p>
      </div>

      <div v-else class="flex flex-col gap-3">
        <div
          v-for="(message, index) in messages"
          :key="index"
          :class="[
            'rounded-lg px-3 py-2 text-sm shadow-sm border',
            message.isUser
              ? 'bg-background border-border ml-auto'
              : 'bg-primary/5 border-primary/20 mr-auto',
          ]"
          style="max-width: 78%"
        >
          <div class="flex items-center gap-2 mb-1">
            <span
              class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium"
              :class="message.isUser ? 'bg-muted text-foreground/80' : 'bg-primary/10 text-primary'"
            >{{ message.isUser ? 'You' : (chatbot?.name || 'AI') }}</span>
            <span class="text-[10px] text-muted-foreground">{{ message.timestamp }}</span>
          </div>
          <div class="whitespace-pre-wrap text-left">
            {{ message.content }}
          </div>
        </div>

        <!-- Typing indicator -->
        <div v-if="isSendingMessage" class="mr-auto flex items-center gap-2 rounded-lg bg-primary/5 px-3 py-2 text-sm border border-primary/20">
          <IconSpinnerArc class="h-4 w-4 animate-spin text-primary" />
          <span class="text-xs text-primary">AI is typing…</span>
        </div>
      </div>
    </div>

    <!-- Input Box -->
    <div class="mt-2">
      <div class="relative">
        <textarea
          v-model="newMessage"
          class="w-full rounded-md border border-input bg-background px-3 py-2 pr-12 resize-none text-sm min-h-[40px] max-h-[96px] shadow-xs focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/50 focus-visible:ring-offset-2"
          rows="1"
          placeholder="Type your message..."
          @keydown.enter.prevent="sendMessage"
        ></textarea>
        <button
          class="absolute right-2 top-2 inline-flex h-7 w-7 items-center justify-center rounded-md text-primary hover:bg-primary/10 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          @click="sendMessage"
          :disabled="isSendingMessage || !newMessage.trim()"
          aria-label="Send message"
        >
          <IconSpinnerArc v-if="isSendingMessage" class="h-4 w-4 animate-spin" />
          <svg v-else xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4">
            <path d="m22 2-7 20-4-9-9-4Z"/>
            <path d="M22 2 11 13"/>
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from "vue";
import { toast } from "vue-sonner";
import IconSpinnerArc from "@/components/icons/IconSpinnerArc.vue";
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
const messagesContainer = ref<HTMLDivElement | null>(null);

const scrollToBottom = async () => {
  await nextTick();
  const el = messagesContainer.value;
  if (el) {
    el.scrollTop = el.scrollHeight;
  }
};

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
  scrollToBottom();

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
      scrollToBottom();
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
    scrollToBottom();
  } finally {
    isSendingMessage.value = false;
    scrollToBottom();
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
  scrollToBottom();
});

// Expose methods for parent component
defineExpose({
  initializeChat,
  resetChat,
  chatbot,
  messages,
});

// Keep pinned to bottom when messages or typing state changes
watch(
  () => [messages.value.length, isSendingMessage.value],
  () => scrollToBottom(),
);
</script>
