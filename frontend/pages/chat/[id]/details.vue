<template>
  <div class="max-w-3xl mx-auto">
    <!-- Loading State -->
    <div v-if="isLoadingChatbot" class="flex items-center justify-center py-12">
      <div
        class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
      ></div>
    </div>

    <!-- Edit Form -->
    <div v-else-if="chatbot">
      <!-- Enable/Disable Toggle -->
      <div class="mb-6 p-4 rounded-lg border border-border bg-card">
        <div class="flex items-center justify-between">
          <div class="space-y-0.5">
            <h3 class="font-medium">Enabled</h3>
            <p class="text-sm text-muted-foreground">
              Toggle chatbot for being enabled or disabled
            </p>
          </div>
          <Switch
            :model-value="chatbot.is_enabled"
            @update:model-value="handleToggleEnabled"
          />
        </div>
      </div>

      <div class="mb-6 rounded-lg border border-border bg-card p-4">
        <div class="flex items-center justify-between">
          <div class="space-y-0.5">
            <h3 class="font-medium">Workspace</h3>
            <p class="text-sm text-muted-foreground">
              {{ chatbot.organization_id ? workspaceName : "Personal workspace" }}
            </p>
          </div>
          <Badge v-if="chatbot.organization_id" variant="secondary">Organization</Badge>
        </div>
        <div
          v-if="!chatbot.organization_id && transferableOrgs.length"
          class="mt-4 grid gap-3 sm:grid-cols-[1fr_auto]"
        >
          <Select v-model="transferOrgId">
            <SelectTrigger class="w-full">
              <SelectValue placeholder="Choose an organization" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem
                v-for="org in transferableOrgs"
                :key="org.id"
                :value="org.id"
              >
                {{ org.name }} ({{ org.role }})
              </SelectItem>
            </SelectContent>
          </Select>
          <Button
            :disabled="!transferOrgId || transferring"
            class="justify-center"
            @click="handleTransfer"
          >
            {{ transferring ? "Moving..." : "Transfer" }}
          </Button>
        </div>
        <p
          v-else-if="!chatbot.organization_id"
          class="mt-3 text-sm text-muted-foreground"
        >
          Join or create an organization to share this chatbot.
        </p>
        <p
          v-else
          class="mt-3 text-sm text-muted-foreground"
        >
          This chatbot is shared with members of {{ workspaceName }}.
        </p>
      </div>

      <ChatbotForm
        mode="edit"
        :chatbot="chatbot"
        :is-loading="isUpdating"
        :shared-knowledge-bases="sharedKnowledgeBases"
        @submit="handleUpdate"
      />
    </div>

    <!-- File Upload Section -->
    <div
      v-if="chatbot"
      id="knowledgeSection"
      class="mt-8 pt-8 border-t border-border"
    >
      <KnowledgeBase :resource-id="chatId" scope="chatbot" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from "vue";
import ChatbotForm from "../components/ChatbotForm.vue";
import KnowledgeBase from "./components/KnowledgeBase.vue";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type {
  ChatbotResponse,
  SharedKnowledgeBaseListResponse,
} from "~/types/api";
import { useRoute } from "vue-router";
import { useApiService } from "@/composables/useApiService";
import { useOrganizations } from "~/composables/useOrganizations";

// Route & API
const route = useRoute();
const apiService = useApiService();
const orgStore = useOrganizations();
const personalId = "00000000-0000-0000-0000-000000000000";
const chatId = computed(() => route.params.id as string);

// State
const chatbot = ref<ChatbotResponse | null>(null);
const isToggling = ref(false);
const transferOrgId = ref<string>("");
const transferring = ref(false);

// Refs
const knowledgeSection = ref<HTMLElement | null>(null);

// API calls
const {
  data,
  execute: executeFetchChatbot,
  error: fetchChatbotError,
  isLoading: isLoadingChatbot,
} = apiService.getChatbot();

const { execute: executeToggle, error: errorToggle } =
  apiService.toggleChatbot();

const {
  execute: executeUpdate,
  error: updateError,
  isLoading: isUpdating,
} = apiService.updateChatbot();
const {
  execute: executeTransfer,
  data: transferData,
  error: transferError,
} = apiService.transferChatbot();

const { execute: loadSharedKnowledgeBases, data: sharedKnowledgeBasesData } =
  apiService.listSharedKnowledgeBases();

const sharedKnowledgeBases = computed(() => {
  const response = sharedKnowledgeBasesData.value as
    | SharedKnowledgeBaseListResponse
    | undefined;
  return response?.knowledge_bases ?? [];
});

const transferableOrgs = computed(() =>
  orgStore.state.value.organizations.filter(
    (o) => o.id !== personalId && ["owner", "admin"].includes(o.role),
  ),
);

const workspaceName = computed(() => {
  if (!chatbot.value?.organization_id) return "Personal";
  const match = orgStore.state.value.organizations.find(
    (o) => o.id === chatbot.value?.organization_id,
  );
  return match?.name ?? "Organization";
});

// Fetch chatbot data
const fetchChatbotData = async () => {
  if (!chatId.value) return;

  await executeFetchChatbot(chatId.value);
  if (fetchChatbotError.value) {
    return;
  }
  if (data.value?.chatbot) {
    chatbot.value = data.value.chatbot;
  }
};

// Handle toggle enabled/disabled
const handleToggleEnabled = async () => {
  if (!chatId.value || !chatbot.value) return;

  isToggling.value = true;
  const newEnabledState = !chatbot.value.is_enabled;

  await executeToggle({
    chatbotId: chatbot.value.id,
    isEnabled: newEnabledState,
  });

  isToggling.value = false;

  if (!errorToggle.value) {
    chatbot.value.is_enabled = newEnabledState;
  }
};

// Handle update
const handleUpdate = async (formData: any) => {
  if (!chatId.value) return;

  await executeUpdate({
    id: chatId.value,
    ...formData,
  });

  if (!updateError.value && chatbot.value) {
    chatbot.value = { ...chatbot.value, ...formData };
  }
};

const handleTransfer = async () => {
  if (!chatbot.value || !transferOrgId.value) return;
  transferring.value = true;
  await executeTransfer({
    chatbotId: chatbot.value.id,
    organizationId: transferOrgId.value,
  });
  transferring.value = false;

  const payload = transferData.value as { chatbot?: ChatbotResponse } | null;
  if (!transferError.value && payload?.chatbot) {
    chatbot.value = payload.chatbot;
    const targetOrg = orgStore.state.value.organizations.find(
      (o) => o.id === transferOrgId.value,
    );
    if (targetOrg) {
      orgStore.setCurrent(targetOrg);
    }
    // Simple flow: switch org and reload chat list under new context
    window.location.assign("/chat");
  }
};

// Scroll to knowledge section
const scrollToKnowledge = async () => {
  await nextTick();
  if (knowledgeSection.value) {
    knowledgeSection.value.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  }
};

// Watch for route changes
watch(
  () => route.params.id,
  (newId) => {
    if (newId) {
      fetchChatbotData();
    }
  },
);

watch(
  transferableOrgs,
  (orgList) => {
    if (!transferOrgId.value && orgList.length) {
      transferOrgId.value = orgList[0].id;
    }
  },
  { immediate: true },
);

watch(
  () => chatbot.value?.organization_id,
  (orgId) => {
    if (!orgId && transferableOrgs.value.length && !transferOrgId.value) {
      transferOrgId.value = transferableOrgs.value[0].id;
    }
  },
);

// Initialize
onMounted(() => {
  fetchChatbotData();
  loadSharedKnowledgeBases();
  orgStore.load();
});

// Expose scroll function for external use
defineExpose({
  scrollToKnowledge,
  knowledgeSection,
});
</script>
