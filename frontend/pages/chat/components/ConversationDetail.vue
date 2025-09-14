<template>
  <div class="h-full flex flex-col">
    <!-- Header -->
    <div class="flex items-center gap-3 pb-4 border-b">
      <Button variant="ghost" size="icon" @click="emit('back')">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="m15 18-6-6 6-6" />
        </svg>
      </Button>
      <div class="flex-1">
        <h3 class="font-medium">Conversation</h3>
        <p class="text-sm text-muted-foreground">
          {{ formatDate(props.conversation?.created_at) }}
        </p>
      </div>
    </div>

    <!-- Messages -->
    <div
      v-if="props.conversation"
      class="flex-1 overflow-y-auto py-4 space-y-4"
    >
      <div
        v-for="(message, index) in props.conversation.messages"
        :key="index"
        :class="[
          'flex',
          message.role === 'user' ? 'justify-end' : 'justify-start',
        ]"
      >
        <div
          :class="[
            'max-w-[80%] rounded-lg px-4 py-2',
            message.role === 'user'
              ? 'bg-primary text-primary-foreground'
              : 'bg-muted',
          ]"
        >
          <div class="text-sm whitespace-pre-wrap">{{ message.content }}</div>
          <div
            :class="[
              'text-xs mt-1',
              message.role === 'user'
                ? 'text-primary-foreground/70'
                : 'text-muted-foreground',
            ]"
          >
            {{ message.timestamp }}
          </div>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-else class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <div
          class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"
        ></div>
        <p class="text-sm text-muted-foreground">Loading conversation...</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import type { Conversation } from "~/types/chat";

interface Props {
  conversation: Conversation | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  back: [];
}>();

const formatDate = (dateString?: string) => {
  if (!dateString) return "";
  const date = new Date(dateString);
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
};

const formatTime = (date: Date) => {
  return new Intl.DateTimeFormat("en-US", {
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
};
</script>
