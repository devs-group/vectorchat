<template>
  <div class="max-w-3xl mx-auto">
    <ChatSectionCard
      title="Knowledge Base"
      subtitle="Add information sources for your assistant to reference"
      color="green"
      :padded="false"
    >
      <template #icon>
        <IconGrid class="h-5 w-5" />
      </template>

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
              :disabled="isDeletingFileState"
            >
              <IconSpinnerArc
                v-if="isDeletingFileState"
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
    </ChatSectionCard>
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
import ChatSectionCard from "@/components/chat/ChatSectionCard.vue";
import { useGlobalState } from "@/composables/useGlobalState";
import { useApiService } from "@/composables/useApiService";
import { useErrorHandler } from "@/composables/useErrorHandler";

interface Props {
  resourceId: string;
  scope?: "chatbot" | "shared";
}

const props = withDefaults(defineProps<Props>(), {
  scope: "chatbot",
});

// API service
const apiService = useApiService();
const { showError, showSuccess } = useErrorHandler();
const isSharedScope = computed(() => props.scope === "shared");

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

// Fetch knowledge base items
const fetchKnowledgeItems = async () => {
  if (!props.resourceId) return;
  isLoadingFiles.value = true;
  try {
    const { data: filesData, execute: executeFetchFiles } =
      isSharedScope.value
        ? apiService.listSharedKnowledgeBaseFiles(props.resourceId)
        : apiService.listChatFiles(props.resourceId);
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
      isSharedScope.value
        ? apiService.listSharedKnowledgeBaseTextSources(props.resourceId)
        : apiService.listTextSources(props.resourceId);
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
    if (!isSharedScope.value) {
      // update global state for toggling the test chat box only for chatbot scope.
      const { hasKnowledgeBaseData } = useGlobalState();
      hasKnowledgeBaseData.value =
        files.value.length > 0 || textSources.value.length > 0;
    }
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
    if (isSharedScope.value) {
      await apiService.uploadSharedKnowledgeBaseFile(props.resourceId, file);
    } else {
      await apiService.uploadFile(props.resourceId, file);
    }
    showSuccess("File uploaded successfully");
    await fetchKnowledgeItems();
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
    if (isSharedScope.value) {
      await apiService.uploadSharedKnowledgeBaseFile(props.resourceId, file);
    } else {
      await apiService.uploadFile(props.resourceId, file);
    }
    showSuccess("File uploaded successfully");
    await fetchKnowledgeItems();
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
const {
  execute: executeDeleteSharedFile,
  isLoading: isDeletingSharedFile,
  error: deleteSharedFileError,
} = apiService.deleteSharedKnowledgeBaseFile();
const isDeletingFileState = computed(() =>
  isSharedScope.value ? isDeletingSharedFile.value : isDeletingFile.value,
);
// Delete a file
const deleteFile = async (filename: string) => {
  if (isSharedScope.value) {
    await executeDeleteSharedFile({ kbId: props.resourceId, filename });
    if (deleteSharedFileError.value) {
      return;
    }
  } else {
    await executeDeleteFile({ chatID: props.resourceId, filename });
    if (deleteFileError.value) {
      return;
    }
  }
  await fetchKnowledgeItems();
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
  execute: uploadChatText,
  error: uploadChatTextError,
  isLoading: isUploadingChatText,
} = apiService.uploadText();
const {
  execute: uploadSharedText,
  error: uploadSharedTextError,
  isLoading: isUploadingSharedText,
} = apiService.uploadSharedKnowledgeBaseText();
const isUploadingText = computed(() =>
  isSharedScope.value ? isUploadingSharedText.value : isUploadingChatText.value,
);

// Add text source (calls backend)
const addTextSource = async () => {
  if (!props.resourceId || !textSource.value.trim()) return;
  if (isSharedScope.value) {
    await uploadSharedText({ kbId: props.resourceId, text: textSource.value.trim() });
    if (uploadSharedTextError.value) {
      return;
    }
  } else {
    await uploadChatText({ chatID: props.resourceId, text: textSource.value.trim() });
    if (uploadChatTextError.value) {
      return;
    }
  }
  textSource.value = "";
  await fetchKnowledgeItems();
};

const { execute: executeDeleteText, error: deleteTextError } =
  apiService.deleteTextSource();
const {
  execute: executeDeleteSharedText,
  error: deleteSharedTextError,
} = apiService.deleteSharedKnowledgeBaseTextSource();
// Delete a text source
const deleteText = async (id: string) => {
  if (isSharedScope.value) {
    await executeDeleteSharedText({ kbId: props.resourceId, id });
    if (deleteSharedTextError.value) {
      return;
    }
  } else {
    await executeDeleteText({ chatID: props.resourceId, id });
    if (deleteTextError.value) {
      return;
    }
  }
  await fetchKnowledgeItems();
};

const {
  execute: uploadWebsite,
  error: uploadWebsiteError,
  isLoading: isIndexingChatWebsite,
} = apiService.uploadWebsite();
const {
  execute: uploadSharedWebsite,
  error: uploadSharedWebsiteError,
  isLoading: isIndexingSharedWebsite,
} = apiService.uploadSharedKnowledgeBaseWebsite();
const isIndexingWebsite = computed(() =>
  isSharedScope.value
    ? isIndexingSharedWebsite.value
    : isIndexingChatWebsite.value,
);

const previewWebsite = () => {
  if (!isWebsiteValid.value || typeof window === "undefined") {
    return;
  }
  const target = normalizedWebsiteUrl.value;
  window.open(target, "_blank", "noopener,noreferrer");
};

// Add website source
const addWebsite = async () => {
  if (!props.resourceId || !isWebsiteValid.value) return;
  const targetUrl = normalizedWebsiteUrl.value;
  indexingTarget.value = targetUrl;
  if (isSharedScope.value) {
    await uploadSharedWebsite({ kbId: props.resourceId, url: targetUrl });
    if (uploadSharedWebsiteError.value) {
      indexingTarget.value = "";
      return;
    }
  } else {
    await uploadWebsite({ chatID: props.resourceId, url: targetUrl });
    if (uploadWebsiteError.value) {
      indexingTarget.value = "";
      return;
    }
  }
  websiteInput.value = "";
  indexingTarget.value = "";
  await fetchKnowledgeItems();
};

// Watch for resource changes
watch(
  () => props.resourceId,
  async (newId) => {
    if (newId) await fetchKnowledgeItems();
  },
);

watch(
  () => props.scope,
  async () => {
    if (props.resourceId) await fetchKnowledgeItems();
  },
);

// Initialize on mount
onMounted(async () => {
  if (props.resourceId) await fetchKnowledgeItems();
});

// Expose methods and reactive state for parent component
defineExpose({ fetchKnowledgeItems, files, textSources });

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
