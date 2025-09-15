<template>
  <div>
    <div class="h-full">
      <div v-if="!selectedConversationId" class="space-y-4">
        <div class="flex justify-end">
          <Button
            variant="ghost"
            size="icon"
            :loading="isLoadingConversationsList"
            aria-label="Refresh conversations"
            @click="refreshConversations"
          >
            <IconRefreshCw class="h-4 w-4" />
            <span class="sr-only">Refresh</span>
          </Button>
        </div>
        <ConversationList
          :conversations="conversations?.conversations ?? null"
          @select="selectConversation"
        />
      </div>
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
import ConversationList from "./components/ConversationList.vue";
import ConversationDetail from "./components/ConversationDetail.vue";
import IconRefreshCw from "@/components/icons/IconRefreshCw.vue";
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

const refreshConversations = async () => {
  await fetchConversationsList({
    chatbotId: chatId.value,
    limit: 20,
    offset: 0,
  });
};

// Watch for selection and load messages on-demand
watch(selectedConversationId, async (sid) => {
  if (!sid) return;
  await fetchMessages({ chatbotId: chatId.value, sessionId: sid });
});
</script>
