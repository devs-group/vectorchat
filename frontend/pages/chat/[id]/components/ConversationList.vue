<template>
  <div class="space-y-4">
    <div v-if="hasAnyConversation" class="space-y-3">
      <div v-if="hasPageConversations" class="space-y-2">
        <div
          v-for="conversation in conversations"
          :key="conversation.session_id"
          @click="emit('select', conversation.session_id)"
          class="p-4 rounded-lg border hover:bg-accent cursor-pointer transition-colors"
        >
          <div class="flex items-start justify-between">
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium truncate">
                {{ getConversationPreview(conversation) }}
              </p>
            </div>
            <div class="flex items-center gap-3">
              <time class="text-xs text-muted-foreground whitespace-nowrap">
                {{ formatRelativeTime(conversation.first_message_at) }}
              </time>
              <button
                type="button"
                class="text-destructive hover:text-destructive/80"
                @click.stop="onDelete(conversation.session_id)"
                aria-label="Delete conversation"
                title="Delete conversation"
              >
                <IconTrash class="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>
      </div>
      <div
        v-else
        class="rounded-lg border border-dashed bg-muted/30 p-6 text-center text-sm text-muted-foreground"
      >
        No conversations on this page
      </div>

      <div v-if="shouldShowPagination" class="pt-2">
        <Pagination
          :items-per-page="itemsPerPage"
          :total="totalItems"
          :page="currentPage"
          @update:page="onPageChange"
        >
          <PaginationContent v-slot="{ items }">
            <PaginationPrevious />

            <template
              v-for="(item, index) in items"
              :key="`${item.type}-${index}-${item.value ?? ''}`"
            >
              <PaginationItem
                v-if="item.type === 'page'"
                :value="item.value"
                :is-active="item.value === currentPage"
                size="default"
              >
                {{ item.value }}
              </PaginationItem>
              <PaginationEllipsis v-else />
            </template>

            <PaginationNext />
          </PaginationContent>
        </Pagination>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="flex flex-col items-center justify-center py-16 px-4">
      <div
        class="w-20 h-20 rounded-full bg-muted/50 flex items-center justify-center mb-6"
      >
        <IconMessageSquareLines
          class="text-muted-foreground"
          width="40"
          height="40"
        />
      </div>

      <h3 class="text-xl font-semibold text-foreground mb-2">
        No conversations yet
      </h3>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import IconMessageSquareLines from "@/components/icons/IconMessageSquareLines.vue";
import IconTrash from "@/components/icons/IconTrash.vue";
import type {
  ConversationListItemResponse,
  ConversationPaginationResponse,
} from "~/types/api";

interface Props {
  conversations: ConversationListItemResponse[] | null;
  pagination?: ConversationPaginationResponse | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  select: [sessionId: string];
  "page-change": [page: number];
  delete: [sessionId: string];
}>();

const DEFAULT_ITEMS_PER_PAGE = 20;

const conversations = computed(() => props.conversations ?? []);
const fallbackPagination = computed<ConversationPaginationResponse>(() => {
  const totalCount = conversations.value.length;
  return {
    page: 1,
    per_page: DEFAULT_ITEMS_PER_PAGE,
    total_items: totalCount,
    total_pages: totalCount > 0 ? 1 : 0,
    has_next_page: false,
    has_prev_page: false,
    offset: 0,
  };
});

const pagination = computed<ConversationPaginationResponse>(() => {
  return props.pagination ?? fallbackPagination.value;
});

const itemsPerPage = computed(() =>
  pagination.value.per_page && pagination.value.per_page > 0
    ? pagination.value.per_page
    : DEFAULT_ITEMS_PER_PAGE,
);

const totalItems = computed(() => pagination.value.total_items ?? 0);
const totalPages = computed(() => pagination.value.total_pages ?? 0);
const currentPage = computed(() => pagination.value.page ?? 1);

const hasAnyConversation = computed(() => totalItems.value > 0);
const hasPageConversations = computed(() => conversations.value.length > 0);
const shouldShowPagination = computed(
	() => hasAnyConversation.value && totalPages.value > 1,
);

const onPageChange = (page: number) => {
  emit("page-change", page);
};

const onDelete = (sessionId: string) => {
  emit("delete", sessionId);
};

const getConversationPreview = (conversation: ConversationListItemResponse) => {
  return conversation.first_message_content || "New conversation";
};

const formatRelativeTime = (dateString: string) => {
  const date = new Date(dateString);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) return `${days}d ago`;
  if (hours > 0) return `${hours}h ago`;
  if (minutes > 0) return `${minutes}m ago`;
  return "Just now";
};
</script>
