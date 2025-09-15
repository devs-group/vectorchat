<template>
  <div>
    <div class="h-full">
      <ConversationList
        v-if="!selectedConversationId"
        :conversations="conversations?.conversations ?? null"
        @select="selectConversation"
      />
      <ConversationDetail
        v-else
        :messages="messages?.messages ?? null"
        @back="handleBack"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import ConversationList from "../components/ConversationList.vue";
import ConversationDetail from "../components/ConversationDetail.vue";
import { useRoute } from "vue-router";
import { useApiService } from "~/composables/useApiService";

// Route
const route = useRoute();
const chatId = computed(() => route.params.id as string);

// State
const selectedConversationId = ref<string | null>(null);
const { listConversations, getConversationMessages } = useApiService();

const {
  data: conversations,
  isLoading: isLoadingConversationsList,
  execute: fetchConversationsList,
} = listConversations();

const {
  data: messages,
  isLoading: isLoadingMessages,
  execute: fetchMessages,
} = getConversationMessages();

onMounted(async () => {
  await fetchConversationsList({
    chatbotId: chatId.value,
    limit: 20,
    offset: 0,
  });
});

// Methods
const selectConversation = (sessionId: string) => {
  selectedConversationId.value = sessionId;
};

const handleBack = () => {
  selectedConversationId.value = null;
};

// Watch for selection and load messages on-demand
watch(selectedConversationId, async (sid) => {
  if (!sid) return;
  await fetchMessages({ chatbotId: chatId.value, sessionId: sid });
});
</script>
