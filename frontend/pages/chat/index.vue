<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <h1 class="text-3xl font-bold tracking-tight">Chats</h1>
      <Button
        v-if="hasChats && !isLoadingChatbots"
        @click="createNewChat"
        class="transition-all hover:shadow-md"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="mr-2 h-4 w-4"
        >
          <path d="M5 12h14"></path>
          <path d="M12 5v14"></path>
        </svg>
        New Chat
      </Button>
    </div>

    <div v-if="isLoadingChatbots" class="flex justify-center py-8">
      <div
        class="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"
      ></div>
    </div>

    <div
      v-else-if="!hasChats"
      class="flex flex-col items-center justify-center p-10 rounded-xl border border-dashed text-center bg-card"
    >
      <div
        class="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-500 text-white shadow-sm mb-4"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 24 24"
          class="h-6 w-6"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
        </svg>
      </div>
      <h3 class="font-medium text-lg mb-1">No chats yet</h3>
      <p class="text-muted-foreground mb-5 max-w-sm">
        Create your first AI assistant to get started.
      </p>
      <Button @click="createNewChat" class="transition-all hover:shadow-sm">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="mr-2 h-4 w-4"
        >
          <path d="M5 12h14"></path>
          <path d="M12 5v14"></path>
        </svg>
        Create New Chat
      </Button>
    </div>

    <div v-else class="grid gap-6 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="chat in data?.chatbots"
        :key="chat.id"
        class="group relative rounded-lg border p-6 hover:border-primary hover:shadow-sm transition-all duration-200 flex flex-col"
      >
        <div class="flex flex-col gap-3">
          <div class="flex items-center justify-between">
            <h3 class="font-semibold text-lg truncate">{{ chat.name }}</h3>
            <div class="flex items-center gap-2">
              <span class="text-xs text-muted-foreground whitespace-nowrap">
                {{ formatDate(chat.created_at) }}
              </span>
              <Button
                variant="ghost"
                size="icon"
                class="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity relative z-10"
                @click.stop="showDeleteConfirmation(chat.id)"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  class="h-4 w-4"
                >
                  <path d="M3 6h18"></path>
                  <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path>
                  <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path>
                </svg>
              </Button>
            </div>
          </div>
          <p class="text-sm text-muted-foreground line-clamp-2">
            {{ chat.description }}
          </p>

          <div class="mt-auto pt-3 border-t flex flex-col gap-1.5">
            <div class="flex items-center gap-2 text-sm text-muted-foreground">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                class="h-4 w-4"
              >
                <path
                  d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"
                ></path>
                <polyline points="13 2 13 9 20 9"></polyline>
              </svg>
              <span>{{ chat.model_name }}</span>
            </div>
            <div class="flex items-center gap-2 text-sm text-muted-foreground">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                class="h-4 w-4"
              >
                <circle cx="12" cy="12" r="10"></circle>
                <polyline points="12 6 12 12 16 14"></polyline>
              </svg>
              <span>Last updated: {{ formatDate(chat.updated_at) }}</span>
            </div>
          </div>
        </div>
        <NuxtLink
          :to="`/chat/${chat.id}`"
          class="absolute inset-0"
          aria-label="View chat"
        ></NuxtLink>
      </div>
    </div>

    <!-- Delete Confirmation Dialog -->
    <Dialog v-model:open="showDeleteDialog">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Delete Chatbot</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete this chatbot? This action cannot be
            undone. All associated files, documents, and conversations will be
            permanently deleted.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter class="gap-2">
          <Button variant="outline" @click="cancelDelete"> Cancel </Button>
          <Button variant="destructive" @click="deleteChat"> Delete </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { toast } from "vue-sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

definePageMeta({
  layout: "authenticated",
});

const apiService = useApiService();

const {
  data,
  execute: listChatbots,
  error: listChatbotsError,
  isLoading: isLoadingChatbots,
} = apiService.listChatbots();

onMounted(async () => {
  try {
    await listChatbots();
    if (data.value?.chatbots && data.value?.chatbots.length > 0) {
      console.log(`Loaded ${data.value?.chatbots.length} chatbots`);
    } else {
      console.log("No chatbots found");
    }
  } catch (error) {
    console.error("Error loading chatbots:", error);
  }
});

// Whether there are any chats
const hasChats = computed(() => (data.value?.chatbots?.length || 0) > 0);

// Format date for display
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

// Create new chat
const router = useRouter();
const createNewChat = async () => {
  router.push("/chat/create");
};

// Delete chat state
const showDeleteDialog = ref(false);
const chatToDelete = ref<string | null>(null);

// Show delete confirmation dialog
const showDeleteConfirmation = (chatId: string) => {
  chatToDelete.value = chatId;
  showDeleteDialog.value = true;
};

// Delete chat
const deleteChat = async () => {
  if (!chatToDelete.value) return;

  try {
    const { execute: executeDelete } = apiService.deleteChatbot(
      chatToDelete.value,
    );
    await executeDelete();

    // Refresh the chatbots list
    await listChatbots();

    toast.success("Chatbot deleted successfully");
  } catch (error) {
    console.error("Error deleting chatbot:", error);
    toast.error("Failed to delete chatbot", {
      description: (error as Error)?.message || "An error occurred",
    });
  } finally {
    showDeleteDialog.value = false;
    chatToDelete.value = null;
  }
};

// Cancel delete
const cancelDelete = () => {
  showDeleteDialog.value = false;
  chatToDelete.value = null;
};
</script>
