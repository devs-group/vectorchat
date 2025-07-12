<template>
  <div class="max-w-lg mx-auto">
    <div class="mb-6">
      <h3 class="text-lg font-semibold mb-2">Files</h3>
      <p class="text-sm text-muted-foreground">
        Upload files for your chatbot to reference during conversations
      </p>
    </div>

    <!-- Upload Button -->
    <div class="mb-4">
      <Button
        @click="handleUploadFile"
        class="transition-all hover:shadow-md"
        variant="outline"
        :disabled="isUploading"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="mr-2 h-4 w-4"
        >
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
          <polyline points="17 8 12 3 7 8"></polyline>
          <line x1="12" y1="3" x2="12" y2="15"></line>
        </svg>
        <span v-if="isUploading">Uploading...</span>
        <span v-else>Upload File</span>
      </Button>
      <input
        type="file"
        ref="fileInput"
        class="hidden"
        @change="onFileSelected"
      />
    </div>

    <!-- Loading State -->
    <div
      v-if="isLoadingFiles"
      class="flex items-center justify-center py-8"
    >
      <div
        class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"
      ></div>
    </div>

    <!-- Files List -->
    <div v-else-if="files.length > 0" class="rounded border p-4 bg-white/60">
      <h4 class="font-medium text-base mb-3">Uploaded Files</h4>
      <div class="space-y-2">
        <div
          v-for="file in files"
          :key="file.filename"
          class="flex items-center justify-between gap-3 rounded border px-3 py-2 bg-white"
        >
          <div class="flex items-center gap-2 flex-1 min-w-0">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="h-4 w-4 text-muted-foreground flex-shrink-0"
            >
              <path
                d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"
              ></path>
              <polyline points="13 2 13 9 20 9"></polyline>
            </svg>
            <div class="flex-1 min-w-0">
              <div class="text-sm font-medium truncate">{{ file.filename }}</div>
              <div class="text-xs text-muted-foreground">
                {{ formatFileSize(file.size) }}
              </div>
            </div>
          </div>
          <Button
            variant="ghost"
            size="sm"
            class="h-8 w-8 p-0"
            @click="deleteFile(file.filename)"
            :disabled="isDeletingFile === file.filename"
          >
            <svg
              v-if="isDeletingFile === file.filename"
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="h-4 w-4 animate-spin"
            >
              <path d="M21 12a9 9 0 11-6.219-8.56"/>
            </svg>
            <svg
              v-else
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="h-4 w-4"
            >
              <path d="M18 6L6 18"></path>
              <path d="M6 6l12 12"></path>
            </svg>
          </Button>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div
      v-else
      class="text-center py-8 rounded border border-dashed bg-muted/20"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="32"
        height="32"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1"
        class="mx-auto mb-2 text-muted-foreground"
      >
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
        <polyline points="17 8 12 3 7 8"></polyline>
        <line x1="12" y1="3" x2="12" y2="15"></line>
      </svg>
      <h4 class="font-medium text-sm mb-1">No files uploaded</h4>
      <p class="text-xs text-muted-foreground">
        Upload files to help your chatbot provide more accurate responses
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { toast } from "vue-sonner";
import { Button } from "@/components/ui/button";
import type { ChatFile } from "~/types/api";

interface Props {
  chatId: string;
}

const props = defineProps<Props>();

// API service
const apiService = useApiService();

// State
const files = ref<ChatFile[]>([]);
const fileInput = ref<HTMLInputElement | null>(null);
const isLoadingFiles = ref(false);
const isUploading = ref(false);
const isDeletingFile = ref<string | null>(null);

// Fetch chat files
const fetchChatFiles = async () => {
  if (!props.chatId) return;

  isLoadingFiles.value = true;

  try {
    const { data: filesData, execute: executeFetchFiles } =
      apiService.listChatFiles(props.chatId);

    await executeFetchFiles();

    if (
      filesData.value &&
      typeof filesData.value === "object" &&
      "files" in filesData.value
    ) {
      files.value = (filesData.value.files as ChatFile[]) || [];
    }
  } catch (error) {
    console.error("Error fetching chat files:", error);
    toast.error("Failed to load files");
  } finally {
    isLoadingFiles.value = false;
  }
};

// Handle file upload button click
const handleUploadFile = () => {
  fileInput.value?.click();
};

// Handle file selection
const onFileSelected = async (event: Event) => {
  const input = event.target as HTMLInputElement;
  if (!input.files || input.files.length === 0) return;

  const file = input.files[0];
  isUploading.value = true;

  try {
    await apiService.uploadFile(props.chatId, file);
    toast.success("File uploaded successfully");
    await fetchChatFiles();
    input.value = "";
  } catch (error) {
    console.error("Error uploading file:", error);
    toast.error("Error uploading file", {
      description: (error as Error)?.message
    });
  } finally {
    isUploading.value = false;
  }
};

// Delete a file
const deleteFile = async (filename: string) => {
  isDeletingFile.value = filename;

  try {
    const { execute: executeDelete } = apiService.deleteFile(props.chatId, filename);
    await executeDelete();

    toast.success("File deleted successfully");
    await fetchChatFiles();
  } catch (error) {
    console.error("Error deleting file:", error);
    toast.error("Error deleting file");
  } finally {
    isDeletingFile.value = null;
  }
};

// Format file size
const formatFileSize = (sizeInBytes: number) => {
  if (sizeInBytes < 1024) {
    return `${sizeInBytes} B`;
  } else if (sizeInBytes < 1024 * 1024) {
    return `${(sizeInBytes / 1024).toFixed(1)} KB`;
  } else {
    return `${(sizeInBytes / (1024 * 1024)).toFixed(1)} MB`;
  }
};

// Watch for chatId changes
watch(
  () => props.chatId,
  async (newChatId) => {
    if (newChatId) {
      await fetchChatFiles();
    }
  },
);

// Initialize on mount
onMounted(async () => {
  if (props.chatId) {
    await fetchChatFiles();
  }
});

// Expose methods for parent component
defineExpose({
  fetchChatFiles,
  files,
});
</script>
