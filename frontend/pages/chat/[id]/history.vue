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
          :pagination="conversations?.pagination ?? null"
          @select="selectConversation"
          @page-change="handlePageChange"
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
const ITEMS_PER_PAGE = 20;
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

const currentPage = computed(() => {
  return conversations.value?.pagination.page ?? 1;
});

const loadConversations = async (page = 1) => {
  const safePage = page < 1 ? 1 : page;
  await fetchConversationsList({
    chatbotId: chatId.value,
    limit: ITEMS_PER_PAGE,
    page: safePage,
  });
};

onMounted(async () => {
  await loadConversations();
});

// Methods
const selectConversation = (sessionId: string) => {
  selectedConversationId.value = sessionId;
};

const handleBack = () => {
  selectedConversationId.value = null;
};

const refreshConversations = async () => {
  await loadConversations(currentPage.value);
};

watch(chatId, async (newId, oldId) => {
  if (!newId || newId === oldId) return;
  selectedConversationId.value = null;
  await loadConversations();
});

const handlePageChange = async (page: number) => {
  await loadConversations(page);
};

// Watch for selection and load messages on-demand
watch(selectedConversationId, async (sid) => {
  if (!sid) return;
  await fetchMessages({ chatbotId: chatId.value, sessionId: sid });
});
</script>
