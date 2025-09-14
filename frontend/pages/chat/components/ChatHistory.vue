<template>
  <div class="max-w-3xl mx-auto">
    <!-- Empty State -->
    <div class="flex flex-col items-center justify-center py-16 px-4">
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
        Oops... No conversations yet
      </h3>

      <p class="text-muted-foreground text-center max-w-md mb-8">
        Start testing your chatbot to see the conversation history here. All
        chat sessions will be saved and displayed in this section.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";

interface Props {
  chatId: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  "switch-to-test": [];
}>();

// TODO: In the future, this component will:
// 1. Fetch conversation history from the API
// 2. Display a list of conversations with timestamps
// 3. Allow viewing individual conversation threads
// 4. Provide filters and search functionality
</script>
