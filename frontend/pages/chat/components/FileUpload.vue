<template>
  <div class="max-w-3xl mx-auto">
    <!-- Card container -->
    <div class="rounded-2xl border border-border bg-card shadow-sm">
      <!-- Header -->
      <div class="px-6 py-5 border-b border-border/70">
        <div class="flex items-start gap-3">
          <div
            class="mt-0.5 inline-flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-emerald-500 to-teal-500 text-white shadow-sm"
          >
            <IconGrid class="h-5 w-5" />
          </div>
          <div>
            <h2 class="text-lg font-medium">Knowledge Base</h2>
            <p class="text-sm text-muted-foreground">
              Add information sources for your assistant to reference
            </p>
          </div>
        </div>
      </div>

      <!-- Tabs -->
      <div class="px-6 pt-6">
        <div class="flex rounded-full bg-muted/60 p-1 text-sm">
          <button
            class="flex-1 inline-flex items-center justify-center gap-2 rounded-full py-2 transition-colors"
            :class="
              activeTab === 'files'
                ? 'bg-background shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            "
            @click="activeTab = 'files'"
          >
            <IconFile class="h-4 w-4" />
            Files
          </button>
          <button
            class="flex-1 inline-flex items-center justify-center gap-2 rounded-full py-2 transition-colors"
            :class="
              activeTab === 'text'
                ? 'bg-background shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            "
            @click="activeTab = 'text'"
          >
            <IconText class="h-4 w-4" />
            Text
          </button>
          <button
            class="flex-1 inline-flex items-center justify-center gap-2 rounded-full py-2 transition-colors"
            :class="
              activeTab === 'website'
                ? 'bg-background shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            "
            @click="activeTab = 'website'"
          >
            <IconGlobe class="h-4 w-4" />
            Website
          </button>
        </div>
      </div>

      <!-- Tab content -->
      <div class="p-6 md:p-8">
        <!-- Files tab -->
        <div v-show="activeTab === 'files'">
          <div
            class="relative rounded-xl border border-dashed border-border/70 bg-muted/20 p-6 md:p-8 text-center"
            @dragover.prevent="isDragging = true"
            @dragleave.prevent="isDragging = false"
            @drop.prevent="handleDrop"
            :class="{
              'ring-2 ring-primary/40 ring-offset-2 ring-offset-background':
                isDragging,
            }"
          >
            <div
              class="mx-auto inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10 text-primary"
            >
              <IconUpload class="h-6 w-6" />
            </div>
            <h4 class="mt-3 text-base font-medium">Upload Files</h4>
            <p class="mt-1 text-xs text-muted-foreground">
              PDF, TXT, DOC, and more (max 10MB each)
            </p>
            <div class="mt-5">
              <Button
                variant="secondary"
                @click="handleUploadFile"
                :disabled="isUploading"
              >
                <span v-if="isUploading">Uploading...</span>
                <span v-else>Choose Files</span>
              </Button>
              <input
                type="file"
                ref="fileInput"
                class="hidden"
                @change="onFileSelected"
              />
            </div>
          </div>
        </div>

        <!-- Text tab -->
        <div v-show="activeTab === 'text'">
          <div>
            <Textarea
              v-model="textSource"
              class="min-h-[120px]"
              placeholder="Paste reference text here"
            />
            <div class="mt-3">
              <Button
                variant="secondary"
                @click="addTextSource"
                :disabled="!textSource.trim()"
                >Add Text</Button
              >
            </div>
          </div>
        </div>

        <!-- Website tab -->
        <div v-show="activeTab === 'website'">
          <div class="max-w-xl">
            <Input
              v-model="websiteUrl"
              :disabled="isIndexingWebsite"
              placeholder="https://docs.company.com"
            />
            <div class="mt-3 flex items-center gap-3">
              <Button
                variant="secondary"
                @click="addWebsite"
                :disabled="!websiteUrl.trim() || isIndexingWebsite"
              >
                <template v-if="isIndexingWebsite">
                  <IconSpinnerArc class="mr-2 h-4 w-4 animate-spin" />
                  Indexing...
                </template>
                <template v-else> Add Website </template>
              </Button>
              <span
                v-if="isIndexingWebsite"
                class="text-xs text-muted-foreground"
                >This may take a minute</span
              >
            </div>
            <div
              v-if="isIndexingWebsite"
              class="mt-3 rounded-lg border bg-muted/30 px-3 py-2 text-xs text-muted-foreground"
            >
              We are crawling pages under your URL and adding them as context.
            </div>
          </div>
        </div>

        <!-- Divider -->
        <div class="my-8 h-px bg-border"></div>

        <!-- Current knowledge sources header -->
        <div class="mb-3 flex items-center justify-between">
          <h4 class="font-medium">Current Knowledge Sources</h4>
          <span
            class="rounded-full bg-muted px-2 py-1 text-xs text-muted-foreground"
            >{{ files.length + textSources.length }} item(s)</span
          >
        </div>

        <!-- Loading -->
        <div
          v-if="isLoadingFiles"
          class="flex items-center justify-center py-8"
        >
          <div
            class="h-6 w-6 animate-spin rounded-full border-b-2 border-primary"
          ></div>
        </div>

        <!-- Empty state -->
        <div
          v-else-if="files.length + textSources.length === 0"
          class="rounded-xl border border-dashed bg-muted/20 p-8 text-center text-sm text-muted-foreground"
        >
          No sources yet — add files, text, or websites above.
        </div>

        <!-- List of files -->
        <div v-else class="space-y-3">
          <div
            v-for="file in files"
            :key="file.filename"
            class="flex items-center justify-between gap-3 rounded-xl border bg-background px-4 py-3 shadow-xs"
          >
            <div class="flex min-w-0 flex-1 items-center gap-3">
              <div
                class="inline-flex h-8 w-8 items-center justify-center rounded-full bg-primary/10 text-primary"
              >
                <IconFile class="h-4 w-4" />
              </div>
              <div class="min-w-0 flex-1">
                <div class="truncate text-sm font-medium">
                  {{ file.filename }}
                </div>
                <div class="text-xs text-muted-foreground">
                  <template
                    v-if="
                      typeof file.size === 'number' && !Number.isNaN(file.size)
                    "
                  >
                    {{ formatFileSize(file.size) }}
                  </template>
                  <template v-else-if="(file as any).uploaded_at">
                    {{ new Date((file as any).uploaded_at).toLocaleString() }}
                  </template>
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
              <IconSpinnerArc
                v-if="isDeletingFile === file.filename"
                class="h-4 w-4 animate-spin"
              />
              <IconX v-else class="h-4 w-4" />
            </Button>
          </div>
          <!-- Text sources list -->
          <div
            v-for="src in textSources"
            :key="src.id"
            class="flex items-center justify-between gap-3 rounded-xl border bg-background px-4 py-3 shadow-xs"
          >
            <div class="flex min-w-0 flex-1 items-center gap-3">
              <div
                class="inline-flex h-8 w-8 items-center justify-center rounded-full bg-primary/10 text-primary"
              >
                <IconText class="h-4 w-4" />
              </div>
              <div class="min-w-0 flex-1">
                <div class="truncate text-sm font-medium">{{ src.title }}</div>
                <div class="text-xs text-muted-foreground">
                  {{ new Date(src.uploaded_at).toLocaleString() }}
                </div>
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              class="h-8 w-8 p-0"
              @click="deleteText(src.id)"
            >
              <IconX class="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { toast } from "vue-sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import type { ChatFile } from "~/types/api";
import IconGrid from "@/components/icons/IconGrid.vue";
import IconFile from "@/components/icons/IconFile.vue";
import IconText from "@/components/icons/IconText.vue";
import IconGlobe from "@/components/icons/IconGlobe.vue";
import IconUpload from "@/components/icons/IconUpload.vue";
import IconSpinnerArc from "@/components/icons/IconSpinnerArc.vue";
import IconX from "@/components/icons/IconX.vue";

interface Props {
  chatId: string;
}

const props = defineProps<Props>();

// API service
const apiService = useApiService();

// State
const files = ref<ChatFile[]>([]);
const textSources = ref<{ id: string; title: string; uploaded_at: string }[]>(
  [],
);
const fileInput = ref<HTMLInputElement | null>(null);
const isLoadingFiles = ref(false);
const isUploading = ref(false);
const isDeletingFile = ref<string | null>(null);
const isDragging = ref(false);
const isIndexingWebsite = ref(false);

// Tabs & inputs
const activeTab = ref<"files" | "text" | "website">("files");
const textSource = ref("");
const websiteUrl = ref("");

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
    const { data: textData, execute: executeFetchTexts } =
      apiService.listTextSources(props.chatId);
    await executeFetchTexts();
    if (
      textData.value &&
      typeof textData.value === "object" &&
      "sources" in textData.value
    ) {
      textSources.value = (textData.value.sources as any[]) || [];
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
      description: (error as Error)?.message,
    });
  } finally {
    isUploading.value = false;
  }
};

// Drag & drop
const handleDrop = async (e: DragEvent) => {
  isDragging.value = false;
  const dt = e.dataTransfer;
  if (!dt || !dt.files || dt.files.length === 0) return;
  const file = dt.files[0];
  isUploading.value = true;
  try {
    await apiService.uploadFile(props.chatId, file);
    toast.success("File uploaded successfully");
    await fetchChatFiles();
  } catch (error) {
    console.error("Error uploading via drop:", error);
    toast.error("Error uploading file");
  } finally {
    isUploading.value = false;
  }
};

// Delete a file
const deleteFile = async (filename: string) => {
  isDeletingFile.value = filename;
  try {
    const { execute: executeDelete } = apiService.deleteFile(
      props.chatId,
      filename,
    );
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
  if (sizeInBytes < 1024) return `${sizeInBytes} B`;
  if (sizeInBytes < 1024 * 1024) return `${(sizeInBytes / 1024).toFixed(1)} KB`;
  return `${(sizeInBytes / (1024 * 1024)).toFixed(1)} MB`;
};

// Add text source (calls backend)
const addTextSource = async () => {
  if (!props.chatId || !textSource.value.trim()) return;
  try {
    const { execute } = apiService.uploadText(
      props.chatId,
      textSource.value.trim(),
    );
    await execute();
    toast.success("Text added successfully");
    textSource.value = "";
    await fetchChatFiles();
  } catch (e) {
    console.error("Error uploading text:", e);
    toast.error("Error uploading text");
  }
};
// Delete a text source
const deleteText = async (id: string) => {
  try {
    const { execute } = apiService.deleteTextSource(props.chatId, id);
    await execute();
    toast.success("Text source deleted successfully");
    await fetchChatFiles();
  } catch (e) {
    console.error("Error deleting text source:", e);
    toast.error("Error deleting text source");
  }
};
const addWebsite = async () => {
  if (!props.chatId || !websiteUrl.value.trim()) return;
  try {
    isIndexingWebsite.value = true;
    const { execute } = apiService.uploadWebsite(
      props.chatId,
      websiteUrl.value.trim(),
    );
    await execute();
    toast.success("Website indexed successfully");
    websiteUrl.value = "";
    await fetchChatFiles();
  } catch (e) {
    console.error("Error adding website:", e);
    toast.error("Error adding website");
  } finally {
    isIndexingWebsite.value = false;
  }
};

// Watch for chatId changes
watch(
  () => props.chatId,
  async (newChatId) => {
    if (newChatId) await fetchChatFiles();
  },
);

// Initialize on mount
onMounted(async () => {
  if (props.chatId) await fetchChatFiles();
});

// Expose methods for parent component
defineExpose({ fetchChatFiles, files });
</script>

<style scoped>
.shadow-xs {
  box-shadow: 0 1px 1px rgba(0, 0, 0, 0.04);
}
</style>
