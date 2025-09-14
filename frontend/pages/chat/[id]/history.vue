<template>
  <div>
    <div class="h-full">
      <ConversationList
        v-if="!selectedConversationId"
        :conversations="conversations"
        @select="selectConversation"
        @switch-to-test="handleSwitchToTest"
      />
      <ConversationDetail
        v-else
        :conversation="selectedConversation"
        @back="handleBack"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import ConversationList from "../components/ConversationList.vue";
import ConversationDetail from "../components/ConversationDetail.vue";
import type { Conversation } from "~/types/chat";
import { useRoute } from "vue-router";

// Route
const route = useRoute();
const chatId = computed(() => route.params.id as string);

// State
const selectedConversationId = ref<string | null>(null);

// Dummy data matching the backend structure
const conversations = ref<Conversation[]>([
  {
    session_id: "550e8400-e29b-41d4-a716-446655440001",
    created_at: new Date(Date.now() - 1000 * 60 * 30).toISOString(), // 30 min ago
    messages: [
      {
        role: "user",
        content: "What are your business hours?",
      },
      {
        role: "assistant",
        content:
          "Our business hours are Monday to Friday, 9 AM to 6 PM EST. We're closed on weekends and public holidays.",
      },
      {
        role: "user",
        content: "Do you offer weekend support?",
      },
      {
        role: "assistant",
        content:
          "We offer emergency support on weekends for premium customers. You can reach our weekend support team at emergency@support.com.",
      },
    ],
  },
  {
    session_id: "550e8400-e29b-41d4-a716-446655440002",
    created_at: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(), // 2 hours ago
    messages: [
      {
        role: "user",
        content: "How do I reset my password?",
      },
      {
        role: "assistant",
        content:
          "To reset your password:\n1. Go to the login page\n2. Click 'Forgot Password'\n3. Enter your email address\n4. Check your email for reset instructions\n5. Follow the link to create a new password",
      },
    ],
  },
  {
    session_id: "550e8400-e29b-41d4-a716-446655440003",
    created_at: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(), // 1 day ago
    messages: [
      {
        role: "user",
        content: "What payment methods do you accept?",
      },
      {
        role: "assistant",
        content:
          "We accept various payment methods including:\n• Credit cards (Visa, MasterCard, Amex)\n• PayPal\n• Bank transfers\n• Cryptocurrency (Bitcoin, Ethereum)",
      },
    ],
  },
]);

// Computed
const selectedConversation = computed<Conversation | null>(() => {
  if (!selectedConversationId.value) return null;
  return (
    conversations.value.find(
      (c) => c.session_id === selectedConversationId.value,
    ) ?? null
  );
});

// Methods
const selectConversation = (sessionId: string) => {
  selectedConversationId.value = sessionId;
};

const handleBack = () => {
  selectedConversationId.value = null;
};

// Handle switch to test (for mobile)
const handleSwitchToTest = () => {
  // This event will bubble up to the parent if needed
  // For now, we just handle it locally
  console.log("Switch to test requested");
};
</script>
