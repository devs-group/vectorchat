<template>
  <div class="space-y-2">
    <!-- Conversations List -->
    <div v-if="props.conversations && props.conversations.length > 0">
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
          <time class="text-xs text-muted-foreground ml-2 whitespace-nowrap">
            {{ formatRelativeTime(conversation.first_message_at) }}
          </time>
        </div>
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
import IconMessageSquareLines from "@/components/icons/IconMessageSquareLines.vue";
import type { ConversationListItemResponse } from "~/types/api";

interface Props {
  conversations: ConversationListItemResponse[] | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  select: [sessionId: string];
}>();

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
