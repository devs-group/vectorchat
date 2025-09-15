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
        <PillTabs v-model="activeTab">
          <PillTab value="files">
            <template #icon>
              <IconFile class="h-4 w-4" />
            </template>
            Files
          </PillTab>
          <PillTab value="text">
            <template #icon>
              <IconText class="h-4 w-4" />
            </template>
            Text
          </PillTab>
          <PillTab value="website">
            <template #icon>
              <IconGlobe class="h-4 w-4" />
            </template>
            Website
          </PillTab>
        </PillTabs>
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
                :disabled="!textSource.trim() || isUploadingText"
                :loading="isUploadingText"
                >Add Text</Button
              >
            </div>
          </div>
        </div>

        <!-- Website tab -->
        <div v-show="activeTab === 'website'">
          <div class="max-w-xl space-y-3">
            <Label for="kb-website-url">Website URL</Label>
            <div class="relative">
              <span
                v-if="websiteProtocolHint"
                class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-2 text-sm text-muted-foreground"
              >
                {{ websiteProtocolHint }}
              </span>
              <Input
                id="kb-website-url"
                v-model="websiteInput"
                :disabled="isIndexingWebsite"
                placeholder="docs.company.com"
                :class="websiteProtocolHint ? 'pl-15 pr-12' : 'pr-12'"
                :aria-invalid="Boolean(websiteError)"
                @keydown.enter.prevent="addWebsite"
              />
              <Button
                v-if="isWebsiteValid"
                type="button"
                variant="ghost"
                size="sm"
                class="absolute right-1 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                @click="previewWebsite"
              >
                <IconGlobe class="h-4 w-4" />
              </Button>
            </div>
            <p
              v-if="websiteError"
              class="flex items-center gap-2 text-xs text-destructive"
            >
              <IconAlertCircle class="h-3.5 w-3.5" />
              {{ websiteError }}
            </p>
            <p
              v-else
              class="flex items-center gap-2 text-xs text-muted-foreground"
            >
              <IconGlobe class="h-3.5 w-3.5" />
              {{ websiteHint }}
            </p>
            <div class="flex items-center gap-3">
              <Button
                variant="secondary"
                @click="addWebsite"
                :disabled="!isWebsiteValid || isIndexingWebsite"
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
                >This may take a moment</span
              >
            </div>
            <div
              v-if="isIndexingWebsite"
              class="mt-3 rounded-lg border bg-muted/30 px-3 py-2 text-xs text-muted-foreground"
            >
              We are crawling pages under
              <span class="font-medium text-foreground">
                {{ indexingTargetDisplay }}
              </span>
              and adding them as context.
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
            >{{ itemsCount }} item(s) • {{ formatFileSize(totalBytes) }}</span
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
                <component
                  :is="
                    file.filename?.startsWith('website-') ? IconGlobe : IconFile
                  "
                  class="h-4 w-4"
                />
              </div>
              <div class="min-w-0 flex-1">
                <div class="truncate text-sm font-medium">
                  {{ file.filename }}
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ formatFileSize(file.size) }} •
                  <span v-if="(file as any).uploaded_at">
                    {{ formatDate((file as any).uploaded_at) }}
                  </span>
                  <span v-else>
                    {{ formatDate((file as any).updated_at) }}
                  </span>
                </div>
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              class="h-8 w-8 p-0"
              @click="deleteFile(file.filename)"
              :disabled="isDeletingFile"
            >
              <IconSpinnerArc
                v-if="isDeletingFile"
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
                  {{ formatFileSize((src as any).size || 0) }} •
                  {{ formatDate(src.uploaded_at) }}
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
import { ref, onMounted, watch, computed } from "vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { PillTabs, PillTab } from "@/components/ui/pill-tabs";
import type { ChatFile } from "~/types/api";
import IconGrid from "@/components/icons/IconGrid.vue";
import IconFile from "@/components/icons/IconFile.vue";
import IconText from "@/components/icons/IconText.vue";
import IconGlobe from "@/components/icons/IconGlobe.vue";
import IconAlertCircle from "@/components/icons/IconAlertCircle.vue";
import IconUpload from "@/components/icons/IconUpload.vue";
import IconSpinnerArc from "@/components/icons/IconSpinnerArc.vue";
import IconX from "@/components/icons/IconX.vue";
import { useGlobalState } from "@/composables/useGlobalState";

interface Props {
  chatId: string;
}

const props = defineProps<Props>();

// API service
const apiService = useApiService();
const { showError, showSuccess } = useErrorHandler();

// State
const files = ref<ChatFile[]>([]);
const textSources = ref<{ id: string; title: string; uploaded_at: string }[]>(
  [],
);
const fileInput = ref<HTMLInputElement | null>(null);
const isLoadingFiles = ref(false);
const isUploading = ref(false);
const isDragging = ref(false);

// Tabs & inputs
const activeTab = ref<"files" | "text" | "website">("files");
const textSource = ref("");
const websiteInput = ref("");
const indexingTarget = ref("");

const websiteProtocolHint = computed(() => {
  const raw = websiteInput.value.trim();
  if (!raw) {
    return "https://";
  }

  const compact = raw.replace(/\s+/g, "");
  const hasProtocol = /^[a-zA-Z][a-zA-Z\d+\-.]*:\/\//.test(compact);
  return hasProtocol ? "" : "https://";
});

const parsedWebsite = computed(() => {
  const raw = websiteInput.value.trim();
  if (!raw) {
    return { url: "", error: "", host: "", path: "", search: "" };
  }

  const compact = raw.replace(/\s+/g, "");
  const hasProtocol = /^[a-zA-Z][a-zA-Z\d+\-.]*:\/\//.test(compact);
  const candidate = hasProtocol ? compact : `https://${compact}`;

  try {
    const url = new URL(candidate);
    if (!url.hostname || !url.hostname.includes(".")) {
      return {
        url: "",
        error: "Enter a full domain like docs.company.com.",
        host: "",
        path: "",
        search: "",
      };
    }
    if (!["http:", "https:"].includes(url.protocol)) {
      return {
        url: "",
        error: "Only HTTP or HTTPS URLs are supported.",
        host: "",
        path: "",
        search: "",
      };
    }

    const host = url.hostname.replace(/^www\./, "");
    const path = url.pathname === "/" ? "" : url.pathname.replace(/\/$/, "");

    return { url: url.toString(), error: "", host, path, search: url.search };
  } catch (error) {
    return {
      url: "",
      error: "Enter a valid website address.",
      host: "",
      path: "",
      search: "",
    };
  }
});

const normalizedWebsiteUrl = computed(() => parsedWebsite.value.url);
const websiteError = computed(() => parsedWebsite.value.error);
const isWebsiteValid = computed(() => Boolean(parsedWebsite.value.url));
const websiteHint = computed(() => {
  if (websiteError.value) {
    return "";
  }

  const defaultHint = "Enter a valid website address.";
  if (!websiteInput.value.trim()) {
    return defaultHint;
  }

  const host = parsedWebsite.value.host;
  const path = parsedWebsite.value.path;
  const search = parsedWebsite.value.search;
  if (!host) {
    return defaultHint;
  }

  return `We'll crawl ${host}${path || ""}${search || ""} and its linked pages.`;
});

const indexingTargetDisplay = computed(() => {
  if (indexingTarget.value) {
    try {
      const url = new URL(indexingTarget.value);
      return url.hostname + url.pathname + url.search;
    } catch (error) {
      return indexingTarget.value;
    }
  }

  if (normalizedWebsiteUrl.value) {
    try {
      const url = new URL(normalizedWebsiteUrl.value);
      return url.hostname + url.pathname + url.search;
    } catch (error) {
      return normalizedWebsiteUrl.value;
    }
  }

  return "your site";
});

// Fetch chat files
const fetchChatFiles = async () => {
  if (!props.chatId) return;
  isLoadingFiles.value = true;
  try {
    const { data: filesData, execute: executeFetchFiles } =
      apiService.listChatFiles(props.chatId);
    await executeFetchFiles();
    if (Array.isArray(filesData.value)) {
      files.value = (filesData.value as ChatFile[]) || [];
    } else if (
      filesData.value &&
      typeof filesData.value === "object" &&
      "files" in filesData.value
    ) {
      files.value = (filesData.value.files as ChatFile[]) || [];
    }

    const { data: textData, execute: executeFetchTexts } =
      apiService.listTextSources(props.chatId);
    await executeFetchTexts();
    if (Array.isArray(textData.value)) {
      textSources.value = (textData.value as any[]) || [];
    } else if (
      textData.value &&
      typeof textData.value === "object" &&
      "sources" in textData.value
    ) {
      textSources.value = (textData.value.sources as any[]) || [];
    }
    // update global state for toggling the test chat box.
    const { hasKnowledgeBaseData } = useGlobalState();
    hasKnowledgeBaseData.value =
      files.value.length > 0 || textSources.value.length > 0;
  } catch (error) {
    console.error("Error fetching chat files:", error);
    showError(error, "Failed to load files");
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
    showSuccess("File uploaded successfully");
    await fetchChatFiles();
    input.value = "";
  } catch (error) {
    console.error("Error uploading file:", error);
    showError(error);
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
    showSuccess("File uploaded successfully");
    await fetchChatFiles();
  } catch (error) {
    console.error("Error uploading via drop:", error);
    showError(error);
  } finally {
    isUploading.value = false;
  }
};

const {
  execute: executeDeleteFile,
  isLoading: isDeletingFile,
  error: deleteFileError,
} = apiService.deleteFile();
// Delete a file
const deleteFile = async (filename: string) => {
  await executeDeleteFile({ chatID: props.chatId, filename });
  if (deleteFileError.value) {
    return;
  }
  await fetchChatFiles();
};

// Format file size
const formatFileSize = (sizeInBytes: number) => {
  if (sizeInBytes < 1024) return `${sizeInBytes} B`;
  if (sizeInBytes < 1024 * 1024) return `${(sizeInBytes / 1024).toFixed(1)} KB`;
  return `${(sizeInBytes / (1024 * 1024)).toFixed(1)} MB`;
};

// Format date consistently
const formatDate = (iso: string | Date) => {
  const d = typeof iso === "string" ? new Date(iso) : iso;
  if (!d || isNaN(d.getTime())) return "";
  return d.toLocaleString();
};

const {
  execute: uploadText,
  error: uploadTextError,
  isLoading: isUploadingText,
} = apiService.uploadText();

// Add text source (calls backend)
const addTextSource = async () => {
  if (!props.chatId || !textSource.value.trim()) return;
  await uploadText({ chatID: props.chatId, text: textSource.value.trim() });
  if (uploadTextError.value) {
    return;
  }
  textSource.value = "";
  await fetchChatFiles();
};

const { execute: executeDeleteText, error: deleteTextError } =
  apiService.deleteTextSource();
// Delete a text source
const deleteText = async (id: string) => {
  await executeDeleteText({ chatID: props.chatId, id });
  if (deleteTextError.value) {
    return;
  }
  await fetchChatFiles();
};

const {
  execute: uploadWebsite,
  error: uploadWebsiteError,
  isLoading: isIndexingWebsite,
} = apiService.uploadWebsite();

const previewWebsite = () => {
  if (!isWebsiteValid.value || typeof window === "undefined") {
    return;
  }
  const target = normalizedWebsiteUrl.value;
  window.open(target, "_blank", "noopener,noreferrer");
};

// Add website source
const addWebsite = async () => {
  if (!props.chatId || !isWebsiteValid.value) return;
  const targetUrl = normalizedWebsiteUrl.value;
  indexingTarget.value = targetUrl;
  await uploadWebsite({ chatID: props.chatId, url: targetUrl });
  if (uploadWebsiteError.value) {
    indexingTarget.value = "";
    return;
  }
  websiteInput.value = "";
  indexingTarget.value = "";
  await fetchChatFiles();
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

// Expose methods and reactive state for parent component
defineExpose({ fetchChatFiles, files, textSources });

// Summary: items and total usage
const itemsCount = computed(
  () => (files.value?.length || 0) + (textSources.value?.length || 0),
);
const totalBytes = computed(() => {
  const filesBytes = (files.value || []).reduce(
    (acc, f) => acc + (typeof f.size === "number" ? f.size : 0),
    0,
  );
  const textBytes = (textSources.value || []).reduce(
    (acc, s: any) => acc + (typeof s.size === "number" ? s.size : 0),
    0,
  );
  return filesBytes + textBytes;
});
</script>

<style scoped>
.shadow-xs {
  box-shadow: 0 1px 1px rgba(0, 0, 0, 0.04);
}
</style>
