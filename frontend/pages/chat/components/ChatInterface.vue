<template>
  <div class="flex flex-col w-full max-w-xl justify-start">
    <!-- Header -->
    <div class="flex items-center gap-2 mb-2">
      <button
        @click="resetChat"
        class="ml-auto text-xs bg-secondary text-secondary-foreground hover:bg-secondary/80 px-2 py-1 rounded-md"
      >
        New Chat
      </button>
      <span
        v-if="props.chatbot && !props.chatbot.is_enabled"
        class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400"
      >
        Disabled
      </span>
    </div>
    <p
      v-if="props.chatbot"
      class="text-muted-foreground text-sm mb-4 text-left"
    >
      {{ props.chatbot.description }}
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
        <div
          class="mx-auto inline-flex h-10 w-10 items-center justify-center rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-500 text-white shadow-sm"
        >
          <IconMessageSquare class="h-5 w-5" />
        </div>
        <h3 class="mt-3 font-medium text-base">No messages yet</h3>
        <p class="text-xs text-muted-foreground mt-1">
          Start a conversation with your AI assistant
        </p>
      </div>

      <div v-else class="flex flex-col gap-3">
        <div
          v-for="(message, index) in messages"
          :key="index"
          :class="[
            'rounded-lg px-3 py-2 text-sm shadow-sm border',
            message.isUser
              ? 'bg-background border-border ml-auto'
              : 'bg-primary/5 border-primary/20 mr-auto flex flex-col gap-2',
          ]"
          style="max-width: 78%"
        >
          <div class="flex items-center gap-2 mb-1">
            <span
              class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium"
              :class="
                message.isUser
                  ? 'bg-muted text-foreground/80'
                  : 'bg-primary/10 text-primary'
              "
              >{{ message.isUser ? "You" : chatbot?.name || "AI" }}</span
            >
            <span class="text-[10px] text-muted-foreground">{{
              message.timestamp
            }}</span>
          </div>
          <div class="text-left flex items-start gap-2">
            <IconSpinnerArc
              v-if="!message.isUser && message.isStreaming"
              class="h-4 w-4 text-primary animate-spin"
            />
            <div class="markdown-content">
              <VueMarkdown
                :source="message.content"
                :options="markdownOptions"
              />
            </div>
          </div>
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
          :disabled="props.chatbot?.is_enabled === false"
        ></textarea>
        <button
          class="absolute right-2 top-2 inline-flex h-7 w-7 items-center justify-center rounded-md text-primary hover:bg-primary/10 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          @click="sendMessage"
          :disabled="
            isSendingMessage ||
            !newMessage.trim() ||
            !!(props.chatbot && !props.chatbot?.is_enabled)
          "
          aria-label="Send message"
        >
          <IconSpinnerArc
            v-if="isSendingMessage"
            class="h-4 w-4 animate-spin"
          />
          <IconSend v-else class="h-4 w-4" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick, onBeforeUnmount } from "vue";

import IconMessageSquare from "@/components/icons/IconMessageSquare.vue";
import IconSend from "@/components/icons/IconSend.vue";
import IconSpinnerArc from "@/components/icons/IconSpinnerArc.vue";
import VueMarkdown from "vue-markdown-render";
import type { ChatbotResponse } from "~/types/api";

interface Props {
  chatbot: ChatbotResponse | null;
}

interface Message {
  content: string;
  isUser: boolean;
  timestamp: string;
  isStreaming?: boolean;
}

const props = defineProps<Props>();

// API service
const apiService = useApiService();
const { showError } = useErrorHandler();

// Chat data
const chatbot = ref<ChatbotResponse | null>(null);
const messages = ref<Message[]>([]);
const newMessage = ref("");
const sessionId = ref<string | null>(null);

// Loading state
const messagesContainer = ref<HTMLDivElement | null>(null);
const isSendingMessage = ref(false);
let cancelStream: (() => void) | null = null;

const scrollToBottom = async () => {
  await nextTick();
  const el = messagesContainer.value;
  if (el) {
    el.scrollTop = el.scrollHeight;
  }
};

const markdownOptions = { breaks: true, linkify: true };

// Send a message
const sendMessage = async () => {
  if (!newMessage.value.trim() || isSendingMessage.value) return;

  const userMessage = newMessage.value.trim();
  messages.value.push({
    content: userMessage,
    isUser: true,
    timestamp: new Date().toLocaleTimeString(),
  });

  newMessage.value = "";
  await scrollToBottom();

  if (!props.chatbot?.id) {
    const message = "Chatbot configuration is missing. Please refresh the page.";
    showError(message);
    messages.value.push({
      content: message,
      isUser: false,
      timestamp: new Date().toLocaleTimeString(),
    });
    await scrollToBottom();
    return;
  }

  const assistantMessage: Message = {
    content: "",
    isUser: false,
    timestamp: new Date().toLocaleTimeString(),
    isStreaming: true,
  };

  messages.value.push(assistantMessage);
  await scrollToBottom();
  isSendingMessage.value = true;

  cancelStream = apiService.streamChatMessage(
    {
      chatID: props.chatbot.id,
      query: userMessage,
      sessionId: sessionId.value,
    },
    {
      onChunk: async (chunk) => {
        assistantMessage.content += chunk;
        messages.value = [...messages.value];
        await scrollToBottom();
      },
      onDone: async ({ content, sessionId: newSessionId }) => {
        assistantMessage.content = content;
        if (newSessionId) {
          sessionId.value = newSessionId;
        }
        assistantMessage.isStreaming = false;
        messages.value = [...messages.value];
        isSendingMessage.value = false;
        cancelStream = null;
        await scrollToBottom();
      },
      onError: async ({ message }) => {
        assistantMessage.content = message;
        assistantMessage.isStreaming = false;
        messages.value = [...messages.value];
        showError(message);
        isSendingMessage.value = false;
        cancelStream = null;
        await scrollToBottom();
      },
    },
  );
};

// Reset chat data
const resetChat = () => {
  if (cancelStream) {
    cancelStream();
    cancelStream = null;
  }
  messages.value = [];
  newMessage.value = "";
  sessionId.value = null;
  isSendingMessage.value = false;
};

// Initialize on mount
onMounted(async () => {
  scrollToBottom();
});

onBeforeUnmount(() => {
  if (cancelStream) {
    cancelStream();
    cancelStream = null;
  }
});

// Expose methods for parent component
defineExpose({
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

<style scoped>
.markdown-content {
  width: 100%;
  word-break: break-word;
}

.markdown-content :deep(p) {
  margin: 0;
}

.markdown-content :deep(pre) {
  margin: 0;
  overflow-x: auto;
}

.markdown-content :deep(code) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,
    "Liberation Mono", "Courier New", monospace;
}
</style>
