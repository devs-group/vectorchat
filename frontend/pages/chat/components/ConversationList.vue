<template>
  <div class="space-y-2">
    <!-- Conversations List -->
    <div v-if="conversations.length > 0">
      <div
        v-for="conversation in conversations"
        :key="conversation.session_id"
        @click="$emit('select', conversation.session_id)"
        class="p-4 rounded-lg border hover:bg-accent cursor-pointer transition-colors"
      >
        <div class="flex items-start justify-between">
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium truncate">
              {{ getConversationPreview(conversation) }}
            </p>
            <p class="text-xs text-muted-foreground mt-1">
              {{ conversation.messages.length }} messages
            </p>
          </div>
          <time class="text-xs text-muted-foreground ml-2 whitespace-nowrap">
            {{ formatRelativeTime(conversation.created_at) }}
          </time>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="flex flex-col items-center justify-center py-16 px-4">
      <div
        class="w-20 h-20 rounded-full bg-muted/50 flex items-center justify-center mb-6"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="40"
          height="40"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="text-muted-foreground"
        >
          <path
            d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"
          />
          <line x1="9" y1="10" x2="15" y2="10" />
          <line x1="12" y1="13" x2="12" y2="13" />
        </svg>
      </div>

      <h3 class="text-xl font-semibold text-foreground mb-2">
        No conversations yet
      </h3>

      <p class="text-muted-foreground text-center max-w-md mb-8">
        Start testing your chatbot to see the conversation history here.
      </p>

      <Button @click="$emit('switch-to-test')">Start a Conversation</Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import type { Conversation } from "~/types/chat";

interface Props {
  conversations: Conversation[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  select: [sessionId: string];
  "switch-to-test": [];
}>();

const getConversationPreview = (conversation: Conversation) => {
  const firstUserMessage = conversation.messages.find((m) => m.role === "user");
  return firstUserMessage?.content || "New conversation";
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
