<template>
  <div class="max-w-3xl mx-auto">
    <!-- Loading State -->
    <div v-if="isLoadingChatbot" class="flex items-center justify-center py-12">
      <div
        class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
      ></div>
    </div>

    <!-- Error State -->
    <div v-else-if="chatbotError" class="text-center py-12">
      <div class="text-destructive mb-4">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="48"
          height="48"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="mx-auto mb-2"
        >
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="15" y1="9" x2="9" y2="15"></line>
          <line x1="9" y1="9" x2="15" y2="15"></line>
        </svg>
      </div>
      <h3 class="text-lg font-medium mb-2">Error Loading Chatbot</h3>
      <p class="text-muted-foreground mb-4">{{ chatbotError.message }}</p>
      <Button @click="$emit('retry')" variant="outline">Try Again</Button>
    </div>

    <!-- Edit Form -->
    <div v-else-if="chatbot">
      <!-- Enable/Disable Toggle -->
      <div class="mb-6 p-4 rounded-lg border border-border bg-card">
        <div class="flex items-center justify-between">
          <div class="space-y-0.5">
            <h3 class="font-medium">Enabled</h3>
          </div>
          <button
            @click="$emit('toggle-enabled')"
            :disabled="isToggling"
            :class="[
              'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
              chatbot.is_enabled ? 'bg-primary' : 'bg-muted',
              isToggling ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer',
            ]"
          >
            <span
              :class="[
                'inline-block h-4 w-4 transform rounded-full bg-white transition-transform',
                chatbot.is_enabled ? 'translate-x-6' : 'translate-x-1',
              ]"
            />
          </button>
        </div>
      </div>

      <ChatbotForm
        mode="edit"
        :chatbot="chatbot"
        :is-loading="isUpdating"
        @submit="$emit('update', $event)"
      />
    </div>

    <!-- File Upload Section -->
    <div
      v-if="chatbot"
      ref="knowledgeSection"
      class="mt-8 pt-8 border-t border-border"
    >
      <FileUpload :chat-id="chatId" ref="fileUpload" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import ChatbotForm from "./ChatbotForm.vue";
import FileUpload from "./FileUpload.vue";
import { Button } from "@/components/ui/button";
import type { ChatbotResponse } from "~/types/api";

interface Props {
  chatbot: ChatbotResponse | null;
  chatId: string;
  isLoadingChatbot: boolean;
  chatbotError: Error | null;
  isToggling: boolean;
  isUpdating: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  "toggle-enabled": [];
  update: [formData: any];
  retry: [];
  "knowledge-update": [hasKnowledge: boolean];
}>();

const knowledgeSection = ref<HTMLElement | null>(null);
const fileUpload = ref<InstanceType<typeof FileUpload> | null>(null);

// Watch for knowledge base changes
watch(
  () => [
    fileUpload.value?.files?.length,
    (fileUpload.value as any)?.textSources?.length,
  ],
  () => {
    const filesLen = Array.isArray(fileUpload.value?.files)
      ? (fileUpload.value!.files as any[]).length
      : 0;
    const textLen = Array.isArray((fileUpload.value as any)?.textSources)
      ? ((fileUpload.value as any).textSources as any[]).length
      : 0;
    emit("knowledge-update", filesLen + textLen > 0);
  },
  { immediate: true },
);

// Expose the knowledge section for scrolling
defineExpose({
  knowledgeSection,
  fileUpload,
});
</script>
