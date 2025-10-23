<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from "vue";
import { useRoute, useRouter, useRuntimeConfig } from "#imports";
import { Button } from "@/components/ui/button";
import { useKratosSession } from "@/composables/useKratosSession";

const route = useRoute();
const router = useRouter();
const config = useRuntimeConfig();
const { session, loadSession } = useKratosSession();
const loginHref = ref<string>(config.public.frontendLoginUrl || "#");
const isCheckingSession = ref(true);
const shouldShowLoginPrompt = ref(false);
const isAuthenticated = computed(() => Boolean(session.value));
const isInteractionDisabled = computed(
  () => shouldShowLoginPrompt.value || isCheckingSession.value,
);

const chatbotId = route.params.id as string;
const siteUrl = ref((route.query.siteUrl as string) || "");
const isLoading = ref(true);
const error = ref("");
const chatbotData = ref<any>(null);

// Chat interface state
const messages = ref<
  Array<{ role: "user" | "assistant"; content: string; timestamp: Date }>
>([]);
const currentMessage = ref("");
const isSending = ref(false);
const chatContainer = ref<HTMLElement>();
const sessionId = ref(undefined);

async function sendMessage() {
  if (isInteractionDisabled.value) {
    shouldShowLoginPrompt.value = !isAuthenticated.value;
    return;
  }

  if (!currentMessage.value.trim() || isSending.value) return;

  const userMessage = currentMessage.value.trim();
  currentMessage.value = "";
  isSending.value = true;

  // Add user message
  messages.value.push({
    role: "user",
    content: userMessage,
    timestamp: new Date(),
  });

  // Scroll to bottom
  await nextTick();
  chatContainer.value?.scrollTo({
    top: chatContainer.value.scrollHeight,
    behavior: "smooth",
  });

  try {
    // Send message to chatbot
    const response = await fetch(`/api/chatbot/${chatbotId}/message`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        query: userMessage,
        session_id: sessionId.value,
      }),
    });

    if (response.ok) {
      const data = await response.json();
      sessionId.value = data.session_id;
      messages.value.push({
        role: "assistant",
        content:
          data.response ||
          "I apologize, but I couldn't generate a response. Please try again.",
        timestamp: new Date(),
      });
    } else {
      throw new Error("Failed to get response");
    }
  } catch (err) {
    console.error("Error sending message:", err);
    messages.value.push({
      role: "assistant",
      content:
        "I'm sorry, I'm having trouble responding right now. Please try again in a moment.",
      timestamp: new Date(),
    });
  } finally {
    isSending.value = false;
    await nextTick();
    chatContainer.value?.scrollTo({
      top: chatContainer.value.scrollHeight,
      behavior: "smooth",
    });
  }
}

function formatTime(date: Date) {
  return date.toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  });
}

function goBack() {
  router.push("/");
}

const updateLoginHref = () => {
  if (typeof window === "undefined" || !config.public.frontendLoginUrl) {
    if (!loginHref.value) {
      loginHref.value = "#";
    }
    return;
  }

  try {
    const url = new URL(config.public.frontendLoginUrl);
    url.searchParams.set("return_to", window.location.href);
    loginHref.value = url.toString();
  } catch (error) {
    console.warn(
      "Failed to construct login URL for preview login prompt",
      error,
    );
    loginHref.value = config.public.frontendLoginUrl || "#";
  }

  if (!loginHref.value) {
    loginHref.value = "#";
  }
};

const refreshSession = async () => {
  isCheckingSession.value = true;
  try {
    await loadSession();
  } finally {
    isCheckingSession.value = false;
    shouldShowLoginPrompt.value = !session.value;
  }
};

const handleFocus = async () => {
  await refreshSession();
};

const handleVisibilityChange = async () => {
  if (document.visibilityState === "visible") {
    await refreshSession();
  }
};

onMounted(async () => {
  if (!chatbotId) {
    router.push("/");
    return;
  }

  if (typeof window !== "undefined") {
    updateLoginHref();
    await refreshSession();
    window.addEventListener("focus", handleFocus);
    document.addEventListener("visibilitychange", handleVisibilityChange);
  } else {
    await refreshSession();
  }

  try {
    const response = await fetch(`/api/chatbot/${chatbotId}`);

    if (response.ok) {
      chatbotData.value = await response.json();
    } else {
      throw new Error("Failed to load chatbot");
    }
  } catch (err) {
    console.error("Error loading chatbot:", err);
    error.value = "Failed to load chatbot. It may still be processing.";
  } finally {
    isLoading.value = false;
  }

  messages.value.push({
    role: "assistant",
    content: `Hello! I'm your AI assistant for ${siteUrl.value || "this website"}. I've been trained on the website's content and I'm here to help you find information, answer questions, and guide you through the site. What would you like to know?`,
    timestamp: new Date(),
  });
});

onBeforeUnmount(() => {
  if (typeof window === "undefined") return;
  window.removeEventListener("focus", handleFocus);
  document.removeEventListener("visibilitychange", handleVisibilityChange);
});
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
    <Teleport to="body">
      <div
        v-if="shouldShowLoginPrompt"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4 backdrop-blur-sm"
      >
        <div
          class="w-full max-w-md rounded-3xl bg-white p-8 text-slate-900 shadow-2xl shadow-blue-500/10"
        >
          <div v-if="isCheckingSession" class="space-y-4 text-center">
            <div
              class="mx-auto flex size-12 items-center justify-center rounded-full border-4 border-blue-200 border-t-blue-500 animate-spin"
            ></div>
            <p class="text-sm text-slate-500">Checking your account...</p>
          </div>
          <div v-else class="space-y-6">
            <div class="space-y-2 text-center sm:text-left">
              <h2 class="text-2xl font-semibold text-slate-900">
                Sign in to test your chatbot
              </h2>
              <p class="text-sm text-slate-600">
                You need to be signed in to interact with the live preview.
                We'll bring you right back here once you log in.
              </p>
            </div>
            <div
              class="flex flex-col gap-3 sm:flex-row sm:justify-end sm:text-left"
            >
              <Button as="a" :href="loginHref" class="w-full"> Sign in </Button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
    <!-- Header -->
    <header class="bg-white shadow-sm border-b">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center py-6">
          <div class="flex items-center space-x-4">
            <Button
              variant="outline"
              size="sm"
              @click="goBack"
              class="flex items-center gap-2"
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M10 19l-7-7m0 0l7-7m-7 7h18"
                />
              </svg>
              Back
            </Button>
            <div>
              <h1 class="text-xl font-semibold text-gray-900">
                Chatbot Preview
              </h1>
              <p class="text-sm text-gray-600">
                <span v-if="siteUrl">AI Assistant for {{ siteUrl }}</span>
                <span v-else>AI Website Assistant</span>
              </p>
            </div>
          </div>
          <div class="flex items-center space-x-2">
            <span
              class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800"
            >
              ✨ Live Preview
            </span>
          </div>
        </div>
      </div>
    </header>

    <!-- Loading State -->
    <div v-if="isLoading" class="flex items-center justify-center min-h-[60vh]">
      <div class="text-center">
        <div
          class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"
        ></div>
        <p class="text-gray-600">Loading your chatbot...</p>
      </div>
    </div>

    <!-- Error State -->
    <div
      v-else-if="error"
      class="flex items-center justify-center min-h-[60vh]"
    >
      <div class="text-center max-w-md">
        <div class="text-red-500 mb-4">
          <svg
            class="w-16 h-16 mx-auto"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
            />
          </svg>
        </div>
        <h2 class="text-xl font-semibold text-gray-900 mb-2">
          Something went wrong
        </h2>
        <p class="text-gray-600 mb-6">{{ error }}</p>
        <Button @click="goBack">Go Back</Button>
      </div>
    </div>

    <!-- Chat Interface -->
    <div v-else class="max-w-4xl mx-auto p-4">
      <div class="bg-white rounded-lg shadow-lg overflow-hidden">
        <!-- Chat Header -->
        <div
          class="bg-gradient-to-r from-blue-500 to-purple-600 px-6 py-4 text-white"
        >
          <h2 class="text-lg font-semibold">
            {{ chatbotData?.name || "Website Assistant" }}
          </h2>
          <p class="text-blue-100 text-sm">
            {{ chatbotData?.description || "AI-powered website assistant" }}
          </p>
        </div>

        <!-- Chat Messages -->
        <div ref="chatContainer" class="h-96 overflow-y-auto p-6 space-y-4">
          <div
            v-for="(message, index) in messages"
            :key="index"
            class="flex"
            :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
          >
            <div
              class="max-w-xs lg:max-w-md px-4 py-2 rounded-lg"
              :class="
                message.role === 'user'
                  ? 'bg-blue-500 text-white rounded-br-none'
                  : 'bg-gray-100 text-gray-900 rounded-bl-none'
              "
            >
              <p class="text-sm">{{ message.content }}</p>
              <p
                class="text-xs mt-1 opacity-70"
                :class="
                  message.role === 'user' ? 'text-blue-100' : 'text-gray-500'
                "
              >
                {{ formatTime(message.timestamp) }}
              </p>
            </div>
          </div>

          <!-- Typing indicator -->
          <div v-if="isSending" class="flex justify-start">
            <div
              class="bg-gray-100 text-gray-900 rounded-lg rounded-bl-none px-4 py-2"
            >
              <div class="flex space-x-1">
                <div
                  class="w-2 h-2 bg-gray-400 rounded-full animate-bounce"
                ></div>
                <div
                  class="w-2 h-2 bg-gray-400 rounded-full animate-bounce"
                  style="animation-delay: 0.1s"
                ></div>
                <div
                  class="w-2 h-2 bg-gray-400 rounded-full animate-bounce"
                  style="animation-delay: 0.2s"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Chat Input -->
        <div class="border-t bg-gray-50 px-6 py-4">
          <form @submit.prevent="sendMessage" class="flex space-x-2">
            <input
              v-model="currentMessage"
              type="text"
              placeholder="Ask me anything about the website..."
              :disabled="isSending || isInteractionDisabled"
              class="flex-1 border border-gray-300 rounded-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50"
              @keydown.enter="sendMessage"
            />
            <Button
              type="submit"
              :disabled="
                !currentMessage.trim() || isSending || isInteractionDisabled
              "
              class="px-6 py-2"
            >
              <span v-if="!isSending">Send</span>
              <span v-else class="flex items-center gap-2">
                <div
                  class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"
                ></div>
                Sending...
              </span>
            </Button>
          </form>
          <p class="text-xs text-gray-500 mt-2">
            This is a preview of your AI chatbot. It has been trained on your
            website content.
          </p>
        </div>
      </div>

      <!-- Additional Info -->
      <div class="mt-6 bg-white rounded-lg shadow p-6">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">
          🎉 Your Chatbot is Ready!
        </h3>
        <div class="space-y-3">
          <p class="text-gray-600">
            Your AI chatbot has been successfully created and trained on your
            website content from
            <span class="font-mono text-sm bg-gray-100 px-2 py-1 rounded">{{
              siteUrl
            }}</span>
          </p>
          <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <h4 class="font-medium text-blue-900 mb-2">What's next?</h4>
            <ul class="text-sm text-blue-800 space-y-1">
              <li>• Test the chatbot by asking questions about your website</li>
              <li>
                • The AI can help visitors find information and navigate your
                site
              </li>
              <li>• Responses are based on your actual website content</li>
              <li>• Integration options will be available soon</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Custom scrollbar for chat */
.overflow-y-auto::-webkit-scrollbar {
  width: 6px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: #f1f1f1;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: #a1a1a1;
}
</style>
