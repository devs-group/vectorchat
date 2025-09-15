<template>
  <div class="h-full max-h-[900px] flex flex-col overflow-hidden">
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
      </div>
    </div>

    <!-- Messages -->
    <div
      v-if="props.messages && props.messages.length > 0"
      class="flex-1 min-h-0 overflow-y-auto py-4 space-y-4"
    >
      <div
        v-for="(message, index) in props.messages"
        :key="index"
        :class="[
          'flex',
          message.role === 'user' ? 'justify-end' : 'justify-start',
        ]"
      >
        <div>
          <div
            :class="[
              'max-w-[80%] rounded-lg px-4 py-2 mt-2 relative',
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
              {{ formatDate(message.created_at) }}
            </div>

            <!-- Revise button for assistant messages -->
            <button
              v-if="message.role === 'assistant'"
              class="absolute right-4 -bottom-4 ml-auto text-xs bg-white text-secondary-foreground cursor-pointer px-2 py-1 rounded-md border"
              @click="openRevisionDrawer(index)"
              :title="'Revise answer'"
            >
              <p class="text-xs">Improve Message</p>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div
      v-else
      class="flex-1 min-h-0 flex items-center justify-center overflow-hidden"
    >
      <div class="text-center">
        <div
          class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"
        ></div>
        <p class="text-sm text-muted-foreground">Loading conversation...</p>
      </div>
    </div>
  </div>

  <!-- Drawer: Improve answer -->
  <Drawer v-model:open="drawerOpen">
    <DrawerContent>
      <div class="flex h-full flex-col gap-6">
        <div>
          <h3 class="text-lg font-semibold">Improve answer</h3>
        </div>

        <div class="space-y-4 overflow-auto pr-2">
          <div>
            <Label class="mb-2 block">User message</Label>
            <Textarea
              :model-value="selectedPrevUser?.content || ''"
              disabled
              rows="3"
            />
          </div>
          <div>
            <Label class="mb-2 block">AI response</Label>
            <Textarea
              :model-value="selectedAssistant?.content || ''"
              rows="6"
              class="max-h-[140px]"
            />
          </div>
          <div>
            <Label class="mb-2 block">Revised answer</Label>
            <Textarea v-model="revisedAnswer" rows="6" />
          </div>
          <div>
            <Label class="mb-2 block">Revision reason (optional)</Label>
            <Textarea v-model="revisionReason" rows="3" />
          </div>
        </div>

        <div class="mt-auto flex items-center justify-between gap-2">
          <div>
            <Button
              v-if="existingRevision"
              variant="destructive"
              @click="handleCancelRevision"
              :disabled="actionLoading"
            >
              Cancel revision
            </Button>
          </div>
          <div class="ml-auto flex gap-2">
            <Button
              variant="ghost"
              @click="drawerOpen = false"
              :disabled="actionLoading"
              >Close</Button
            >
            <Button
              @click="handleSaveRevision"
              :disabled="!revisedAnswer || actionLoading"
            >
              {{ existingRevision ? "Update Answer" : "Update Answer" }}
            </Button>
          </div>
        </div>
      </div>
    </DrawerContent>
  </Drawer>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Drawer, DrawerContent } from "@/components/ui/drawer";
import { Pencil } from "lucide-vue-next";
import { ref, computed } from "vue";
import { useApiService } from "~/composables/useApiService";
import type { MessageDetails, RevisionResponse } from "~/types/api";

interface Props {
  messages: MessageDetails[] | null;
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

// Revision drawer state
const drawerOpen = ref(false);
const selectedAssistantIndex = ref<number | null>(null);
const selectedAssistant = computed(() =>
  selectedAssistantIndex.value != null
    ? (props.messages?.[selectedAssistantIndex.value] ?? null)
    : null,
);
const selectedPrevUser = computed(() => {
  if (selectedAssistantIndex.value == null || !props.messages) return null;
  for (let i = selectedAssistantIndex.value - 1; i >= 0; i--) {
    const m = props.messages[i];
    if (m.role === "user") return m;
  }
  return null;
});

const revisedAnswer = ref("");
const revisionReason = ref("");
const existingRevision = ref<RevisionResponse | null>(null);
const actionLoading = ref(false);

const api = useApiService();
const { execute: execGetRevisions, data: revisionsData } = api.getRevisions();
const { execute: execCreateRevision, error: createErr } = api.createRevision();
const { execute: execUpdateRevision, error: updateErr } = api.updateRevision();
const { execute: execDeleteRevision, error: deleteErr } = api.deleteRevision();

const openRevisionDrawer = async (assistantIndex: number) => {
  selectedAssistantIndex.value = assistantIndex;
  revisedAnswer.value = "";
  revisionReason.value = "";
  existingRevision.value = null;

  // Try fetch existing revisions for this chatbot and bind if found
  const chatbotId = props.messages?.[assistantIndex]?.chatbot_id;
  if (chatbotId) {
    await execGetRevisions({ chatbotId });
    const list = (revisionsData.value as any)?.revisions as
      | RevisionResponse[]
      | undefined;
    const msgId = props.messages?.[assistantIndex]?.id;
    if (list && msgId) {
      const found = list.find(
        (r) => r.is_active && r.original_message_id === msgId,
      );
      if (found) {
        existingRevision.value = found;
        // prefill from existing
        revisedAnswer.value = found.revised_answer ?? "";
        revisionReason.value = found.revision_reason ?? "";
      }
    }
  }

  drawerOpen.value = true;
};

const handleSaveRevision = async () => {
  if (
    !selectedAssistant.value ||
    !selectedPrevUser.value ||
    !revisedAnswer.value
  )
    return;
  actionLoading.value = true;
  try {
    if (existingRevision.value) {
      await execUpdateRevision({
        revisionId: existingRevision.value.id,
        body: {
          revised_answer: revisedAnswer.value,
          revision_reason: revisionReason.value || undefined,
        },
      });
    } else {
      await execCreateRevision({
        chatbot_id: selectedAssistant.value.chatbot_id,
        original_message_id: selectedAssistant.value.id,
        question: selectedPrevUser.value.content,
        original_answer: selectedAssistant.value.content,
        revised_answer: revisedAnswer.value,
        revision_reason: revisionReason.value || undefined,
      });
    }
    if (!createErr?.value && !updateErr?.value) {
      drawerOpen.value = false;
    }
  } finally {
    actionLoading.value = false;
  }
};

const handleCancelRevision = async () => {
  if (!existingRevision.value) return;
  actionLoading.value = true;
  try {
    await execDeleteRevision(existingRevision.value.id);
    if (!deleteErr?.value) {
      existingRevision.value = null;
      drawerOpen.value = false;
    }
  } finally {
    actionLoading.value = false;
  }
};
</script>
