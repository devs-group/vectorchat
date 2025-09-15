<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <h1 class="text-3xl font-bold tracking-tight">Chats</h1>
      <Button
        v-if="hasChats && !isLoadingChatbots"
        @click="createNewChat"
        class="transition-all hover:shadow-md"
      >
        <IconPlus class="mr-2 h-4 w-4" />
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
        <IconMessageSquare class="h-6 w-6" />
      </div>
      <h3 class="font-medium text-lg mb-1">No chats yet</h3>
      <p class="text-muted-foreground mb-5 max-w-sm">
        Create your first AI assistant to get started.
      </p>
      <Button @click="createNewChat" class="transition-all hover:shadow-sm">
        <IconPlus class="mr-2 h-4 w-4" />
        Create New Chat
      </Button>
    </div>

    <div v-else class="grid gap-6 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
      <Card
        v-for="chat in data?.chatbots"
        :key="chat.id"
        class="group relative overflow-hidden transition-all duration-200 hover:border-primary hover:shadow-md"
      >
        <CardHeader class="gap-3 pb-0">
          <div class="flex items-start justify-between gap-3">
            <div class="flex min-w-0 items-center gap-2">
              <CardTitle class="text-lg font-semibold truncate">
                {{ chat.name }}
              </CardTitle>
              <span
                v-if="!chat.is_enabled"
                class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium bg-yellow-100 text-yellow-800"
              >
                Disabled
              </span>
            </div>
            <div class="flex items-center gap-2">
              <span class="whitespace-nowrap text-xs text-muted-foreground">
                {{ formatDate(chat.created_at) }}
              </span>
              <Button
                variant="ghost"
                size="icon"
                class="relative z-10 h-8 w-8 opacity-0 transition-opacity group-hover:opacity-100"
                @click.stop="showDeleteConfirmation(chat.id)"
              >
                <IconTrash class="h-4 w-4" />
              </Button>
            </div>
          </div>
          <CardDescription class="line-clamp-2 text-sm text-muted-foreground">
            {{ chat.description }}
          </CardDescription>
        </CardHeader>
        <CardFooter class="flex-col items-start gap-2 border-t pt-4 text-sm text-muted-foreground">
          <div class="flex items-center gap-2">
            <IconMessageSquare class="h-4 w-4" />
            <span>{{ formatMessageCount(chat.ai_messages_amount) }}</span>
          </div>
          <div class="flex items-center gap-2">
            <IconFile class="h-4 w-4" />
            <span>{{ chat.model_name }}</span>
          </div>
          <div class="flex items-center gap-2">
            <IconClock class="h-4 w-4" />
            <span>Last updated: {{ formatDate(chat.updated_at) }}</span>
          </div>
        </CardFooter>
        <NuxtLink
          :to="`/chat/${chat.id}`"
          class="absolute inset-0"
          aria-label="View chat"
        ></NuxtLink>
      </Card>
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
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import IconClock from "@/components/icons/IconClock.vue";
import IconFile from "@/components/icons/IconFile.vue";
import IconMessageSquare from "@/components/icons/IconMessageSquare.vue";
import IconPlus from "@/components/icons/IconPlus.vue";
import IconTrash from "@/components/icons/IconTrash.vue";

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

const formatMessageCount = (count?: number | null) => {
  const value = typeof count === "number" ? count : Number(count ?? 0);
  if (!Number.isFinite(value) || value <= 0) {
    return "0 AI messages";
  }
  return value === 1 ? "1 AI message" : `${value} AI messages`;
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

const { execute: executeDelete } = apiService.deleteChatbot();

// Delete chat
const deleteChat = async () => {
  if (!chatToDelete.value) return;
  await executeDelete(chatToDelete.value);
  await listChatbots();
  showDeleteDialog.value = false;
  chatToDelete.value = null;
};

// Cancel delete
const cancelDelete = () => {
  showDeleteDialog.value = false;
  chatToDelete.value = null;
};
</script>
