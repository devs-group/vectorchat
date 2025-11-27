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
                  @keydown.enter.prevent="handleWebsiteEnter"
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
            <div class="flex items-center gap-3">
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

          <div class="mt-6 space-y-4">
            <div class="rounded-xl border border-border bg-card/70 p-5 shadow-sm space-y-4">
              <div class="flex flex-wrap items-center gap-3">
                <div class="flex h-9 w-9 items-center justify-center rounded-full bg-primary/10 text-primary">
                  <IconClock class="h-4 w-4" />
                </div>
                <div class="flex-1 min-w-[180px]">
                  <p class="text-sm font-medium">Crawl mode</p>
                  <p class="text-xs text-muted-foreground">Use the website URL above; jobs run via the queue.</p>
                </div>
                <div class="flex gap-2 w-full sm:w-auto">
                  <Button
                    class="flex-1 sm:flex-none"
                    variant="outline"
                    size="sm"
                    :class="crawlMode === 'once' ? 'border-primary text-primary' : ''"
                    @click="crawlMode = 'once'"
                  >
                    Once
                  </Button>
                  <Button
                    class="flex-1 sm:flex-none"
                    variant="outline"
                    size="sm"
                    :class="crawlMode === 'recurring' ? 'border-primary text-primary' : ''"
                    @click="crawlMode = 'recurring'"
                  >
                    Recurring
                  </Button>
                </div>
              </div>

              <div v-if="crawlMode === 'recurring'" class="space-y-3">
                <div class="grid gap-3 sm:grid-cols-2">
                  <div>
                    <Label>Cadence</Label>
                    <Select
                      :model-value="scheduleForm.frequency"
                      @update:model-value="(v) => (scheduleForm.frequency = v as any)"
                    >
                      <SelectTrigger class="mt-2 w-full">
                        <SelectValue placeholder="Choose cadence" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="minute">Every minute</SelectItem>
                        <SelectItem value="hourly">Hourly</SelectItem>
                        <SelectItem value="daily">Daily</SelectItem>
                        <SelectItem value="weekly">Weekly (Mon)</SelectItem>
                        <SelectItem value="monthly">Monthly (1st)</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label for="time">Time</Label>
                    <Input
                      id="time"
                      type="time"
                      v-model="scheduleForm.time"
                      class="mt-2"
                    />
                  </div>
                </div>
                <div class="flex flex-col sm:flex-row gap-3">
                  <Button
                    class="sm:w-auto w-full"
                    variant="default"
                    :loading="isSavingSchedule"
                    :disabled="isSavingSchedule || !isWebsiteValid"
                    @click="saveSchedule"
                  >
                    Save & queue
                  </Button>
                  <Button
                    v-if="schedule"
                    class="sm:w-auto w-full"
                    variant="outline"
                    :loading="isDeletingSchedule"
                    :disabled="!schedule || isDeletingSchedule"
                    @click="deleteSchedule"
                  >
                    Delete schedule
                  </Button>
                </div>
              </div>

              <div v-else class="flex flex-col sm:flex-row items-start sm:items-center gap-3">
                <Button
                  class="sm:w-auto w-full"
                  variant="secondary"
                  :loading="isIndexingWebsite"
                  :disabled="isIndexingWebsite || !isWebsiteValid"
                  @click="crawlOnce"
                >
                  Crawl once now
                </Button>
                <p class="text-xs text-muted-foreground">
                  We’ll queue a single crawl immediately.
                </p>
              </div>
            </div>

            <div class="rounded-xl border border-border bg-muted/10 p-5 text-sm space-y-2">
              <div class="flex items-center gap-2">
                <IconRefreshCw class="h-4 w-4 text-primary" />
                <span class="font-medium text-foreground">Queue status</span>
              </div>
              <div class="flex items-center gap-2 text-xs text-muted-foreground">
                <IconSpinnerArc v-if="isIndexingWebsite" class="h-4 w-4 animate-spin" />
                <IconRefreshCw v-else class="h-4 w-4 text-emerald-500" />
                <span>
                  {{ queueStatus || "Idle" }}
                  <span v-if="queueStatusAt"> • {{ formatDate(queueStatusAt) }}</span>
                </span>
              </div>
              <div class="text-xs text-muted-foreground space-y-1">
                <div>Last run: {{ lastRunCopy }}</div>
                <div>Next run: {{ nextRunCopy }}</div>
                <div v-if="schedule?.last_error" class="text-destructive">
                  {{ schedule?.last_error }}
                </div>
                <div v-if="isLoadingSchedule">Loading schedule…</div>
                <div v-if="queueMetrics" class="flex flex-wrap gap-2 pt-1">
                  <span class="rounded bg-muted px-2 py-1">Pending: {{ queueMetrics.consumer?.num_pending ?? 0 }}</span>
                  <span class="rounded bg-muted px-2 py-1">Ack pending: {{ queueMetrics.consumer?.num_ack_pending ?? 0 }}</span>
                  <span class="rounded bg-muted px-2 py-1">Waiting: {{ queueMetrics.consumer?.num_waiting ?? 0 }}</span>
                </div>
              </div>
              <div class="flex gap-2 text-xs">
                <Button size="sm" variant="outline" @click="fetchQueueMetrics">
                  Refresh metrics
                </Button>
              </div>
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
import { ref, onMounted, watch, computed, reactive } from "vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PillTabs, PillTab } from "@/components/ui/pill-tabs";
import type { ChatFile, CrawlSchedule } from "~/types/api";
import IconGrid from "@/components/icons/IconGrid.vue";
import IconFile from "@/components/icons/IconFile.vue";
import IconText from "@/components/icons/IconText.vue";
import IconGlobe from "@/components/icons/IconGlobe.vue";
import IconAlertCircle from "@/components/icons/IconAlertCircle.vue";
import IconUpload from "@/components/icons/IconUpload.vue";
import IconSpinnerArc from "@/components/icons/IconSpinnerArc.vue";
import IconX from "@/components/icons/IconX.vue";
import IconClock from "@/components/icons/IconClock.vue";
import IconRefreshCw from "@/components/icons/IconRefreshCw.vue";
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
const schedule = ref<CrawlSchedule | null>(null);
const isLoadingSchedule = ref(false);
const isSavingSchedule = ref(false);
const isDeletingSchedule = ref(false);
const localTimezone =
  Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
const scheduleForm = reactive({
  frequency: "daily" as "minute" | "hourly" | "daily" | "weekly" | "monthly",
  time: "03:00",
  enabled: true,
});
const crawlMode = ref<"once" | "recurring">("recurring");
const queueStatus = ref<string>("");
const queueStatusAt = ref<string>("");
const queueMetrics = ref<any | null>(null);

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

const isIndexingWebsite = ref(false);
const { execute: enqueueChatOnce } = apiService.crawlChatOnce();
const { execute: enqueueSharedOnce } = apiService.crawlSharedOnce();

const previewWebsite = () => {
  if (!isWebsiteValid.value || typeof window === "undefined") {
    return;
  }
  const target = normalizedWebsiteUrl.value;
  window.open(target, "_blank", "noopener,noreferrer");
};

const parseCron = (expr: string) => {
  const parts = expr.trim().split(/\s+/);
  if (parts.length < 5) return;
  const [min, hour, dom, mon, dow] = parts;
  const time = `${hour.padStart(2, "0")}:${min.padStart(2, "0")}`;
  if (parts[0] !== "*" && parts[1] === "*" && dom === "*" && mon === "*" && dow === "*") {
    return { frequency: "minute" as const, time: "00:00" };
  }
  if (parts[1] !== "*" && dom === "*" && mon === "*" && dow === "*") {
    return { frequency: "hourly" as const, time };
  }
  if (dom !== "*" && dom !== "?") {
    return { frequency: "monthly" as const, time };
  }
  if (dow !== "*" && dow !== "?") {
    return { frequency: "weekly" as const, time };
  }
  return { frequency: "daily" as const, time };
};

const toCron = (
  frequency: "minute" | "hourly" | "daily" | "weekly" | "monthly",
  time: string,
) => {
  const [hour, minute] = (time || "03:00").split(":");
  switch (frequency) {
    case "minute":
      return "* * * * *"; // every minute
    case "hourly":
      return `${minute || "00"} * * * *`; // at minute every hour
    case "weekly":
      return `${minute || "00"} ${hour || "03"} * * 1`; // Monday
    case "monthly":
      return `${minute || "00"} ${hour || "03"} 1 * *`; // 1st of month
    default:
      return `${minute || "00"} ${hour || "03"} * * *`;
  }
};

const loadSchedule = async () => {
  if (!props.resourceId) return;
  isLoadingSchedule.value = true;
  const loader = isSharedScope.value
    ? apiService.listSharedCrawlSchedules()
    : apiService.listChatCrawlSchedules();
  await loader.execute(props.resourceId);
  if (loader.data.value && "schedules" in loader.data.value) {
    const first = loader.data.value.schedules[0];
    if (first) {
      schedule.value = first;
      const parsed = parseCron(first.cron_expr || "");
      if (parsed) {
        scheduleForm.frequency = parsed.frequency;
        scheduleForm.time = parsed.time;
      }
      scheduleForm.timezone = first.timezone || scheduleForm.timezone;
      scheduleForm.enabled = first.enabled;
    }
  }
  isLoadingSchedule.value = false;
};

const fetchQueueMetrics = async () => {
  const { execute, data, error } = apiService.getCrawlQueueMetrics();
  await execute();
  if (!error.value && data.value) {
    queueMetrics.value = (data.value as any).data;
  }
};

const saveSchedule = async () => {
  if (!isWebsiteValid.value) {
    showError(new Error("Enter a valid website URL before scheduling."));
    return;
  }
  if (!scheduleForm.time) {
    scheduleForm.time = "03:00";
  }
  isSavingSchedule.value = true;
  queueStatus.value = "Enqueuing first crawl…";
  queueStatusAt.value = new Date().toISOString();
  const saver = isSharedScope.value
    ? apiService.upsertSharedCrawlSchedule()
    : apiService.upsertChatCrawlSchedule();
  const payload: any = {
    body: {
      url: normalizedWebsiteUrl.value,
      cron_expr: toCron(scheduleForm.frequency, scheduleForm.time),
      timezone: localTimezone,
      enabled: scheduleForm.enabled,
    },
  };
  if (isSharedScope.value) payload.kbId = props.resourceId;
  else payload.chatId = props.resourceId;

  const res = await saver.execute(payload);
  if (!saver.error.value && res) {
    schedule.value = res;
    crawlMode.value = "recurring";
    queueStatus.value = "Queued";
    queueStatusAt.value = new Date().toISOString();
  }
  isSavingSchedule.value = false;
};

const deleteSchedule = async () => {
  if (!schedule.value) return;
  isDeletingSchedule.value = true;
  const deleter = isSharedScope.value
    ? apiService.deleteSharedCrawlSchedule()
    : apiService.deleteChatCrawlSchedule();
  const payload: any = { scheduleId: schedule.value.id };
  if (isSharedScope.value) payload.kbId = props.resourceId;
  else payload.chatId = props.resourceId;
  await deleter.execute(payload);
  if (!deleter.error.value) {
    schedule.value = null;
  }
  isDeletingSchedule.value = false;
};

// One-off crawl via queue
const crawlOnce = async () => {
  if (!props.resourceId || !isWebsiteValid.value) return;
  const targetUrl = normalizedWebsiteUrl.value;
  indexingTarget.value = targetUrl;
  isIndexingWebsite.value = true;
  queueStatus.value = "Queued";
  queueStatusAt.value = new Date().toISOString();
  if (isSharedScope.value) {
    await enqueueSharedOnce({ kbId: props.resourceId, url: targetUrl });
  } else {
    await enqueueChatOnce({ chatId: props.resourceId, url: targetUrl });
  }
  indexingTarget.value = "";
  isIndexingWebsite.value = false;
};

const handleWebsiteEnter = async () => {
  if (crawlMode.value === "recurring") {
    await saveSchedule();
  } else {
    await crawlOnce();
  }
};

// Watch for resource changes
watch(
  () => props.resourceId,
  async (newId) => {
    if (newId) {
      await fetchKnowledgeItems();
      await loadSchedule();
    }
  },
);

watch(
  () => props.scope,
  async () => {
    if (props.resourceId) {
      await fetchKnowledgeItems();
      await loadSchedule();
    }
  },
);

// Initialize on mount
onMounted(async () => {
  if (props.resourceId) {
    await fetchKnowledgeItems();
    await loadSchedule();
  }
  fetchQueueMetrics();
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

const lastRunCopy = computed(() => {
  if (schedule.value?.last_run_at) {
    return formatDate(schedule.value.last_run_at);
  }
  return "Not run yet";
});

const nextRunCopy = computed(() => {
  if (schedule.value?.next_run_at) {
    return formatDate(schedule.value.next_run_at);
  }
  return "Scheduled after next refresh";
});

</script>

<style scoped>
.shadow-xs {
  box-shadow: 0 1px 1px rgba(0, 0, 0, 0.04);
}
</style>
