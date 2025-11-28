<script setup lang="ts">
import { toast } from "vue-sonner";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import SharedKnowledgeSelect from "@/components/chat/SharedKnowledgeSelect.vue";
import type { ChatbotResponse, SharedKnowledgeBase } from "~/types/api";
import IconUserCircle from "@/components/icons/IconUserCircle.vue";
import IconSun from "@/components/icons/IconSun.vue";
import ChatSectionCard from "@/components/chat/ChatSectionCard.vue";
import SystemPromptHelperDialog from "@/components/chat/SystemPromptHelperDialog.vue";
import { useApiService } from "@/composables/useApiService";
import type { LLMModel } from "~/types/api";

interface Props {
  chatbot?: ChatbotResponse | null;
  isLoading?: boolean;
  mode?: "create" | "edit";
  sharedKnowledgeBases?: SharedKnowledgeBase[];
}

interface EmitEvents {
  submit: [data: ChatbotFormData];
}

interface ChatbotFormData {
  name: string;
  description: string;
  model_name: string;
  system_instructions: string;
  max_tokens: number;
  temperature_param: number;
  save_messages: boolean;
  use_max_tokens?: boolean;
  shared_knowledge_base_ids: string[];
}

const props = withDefaults(defineProps<Props>(), {
  chatbot: null,
  isLoading: false,
  mode: "create",
  sharedKnowledgeBases: () => [],
});

const emit = defineEmits<EmitEvents>();

type ModelOption = LLMModel & { hint?: string };
const fallbackModels: ModelOption[] = [
  {
    id: "chat-default",
    label: "Chat Default (GPT-4o Mini)",
    provider: "openai",
    advanced: false,
    hint: "Fast, low-latency general chat via proxy",
  },
  {
    id: "claude-default",
    label: "Claude 3.5 Sonnet",
    provider: "anthropic",
    advanced: true,
    hint: "Advanced reasoning; may consume more credits",
  },
  {
    id: "gpt-4o-mini",
    label: "GPT-4o Mini (Direct)",
    provider: "openai",
    advanced: false,
    hint: "Direct OpenAI fallback if proxy models are unavailable",
  },
  {
    id: "gemini-default",
    label: "Gemini 1.5 Flash",
    provider: "google",
    advanced: true,
    hint: "Google Gemini via proxy; great for long-context reads",
  },
  {
    id: "gpt5-default",
    label: "GPT-5",
    provider: "openai",
    advanced: true,
    hint: "Latest OpenAI flagship; highest quality reasoning",
  },
  {
    id: "gpt5-mini",
    label: "GPT-5 Mini",
    provider: "openai",
    advanced: true,
    hint: "Balanced speed and cost for GPT-5 generation",
  },
  {
    id: "gpt5-nano",
    label: "GPT-5 Nano",
    provider: "openai",
    advanced: false,
    hint: "Fast, budget-friendly GPT-5 tier",
  },
];

const defaultModelId = fallbackModels[0].id;
const api = useApiService();

// Form reactive data
const name = ref("");
const description = ref("");
const systemInstructions = ref("You are a helpful AI assistant");
const modelName = ref(defaultModelId);
const temperatureParam = ref(0.7);
const maxTokens = ref(2000);
const saveMessages = ref(true);
const useMaxTokens = ref(false);
const selectedSharedKnowledgeBaseIds = ref<string[]>([]);
const showPromptHelper = ref(false);
const models = ref<ModelOption[]>([...fallbackModels]);
const modelsLoading = ref(false);
const modelsSource = ref<"fallback" | "api">("fallback");

const formatHint = (model?: Partial<ModelOption>) => {
  if (!model) return "";
  const provider = model.provider
    ? `${model.provider.slice(0, 1).toUpperCase()}${model.provider.slice(1)}`
    : "LLM";
  if (model.advanced) {
    return `${provider} • Advanced`;
  }
  return provider;
};

const normalizeModels = (items: LLMModel[]): ModelOption[] => {
  return items.map((m) => ({
    ...m,
    label: m.label || m.id,
    hint: formatHint(m),
  }));
};

const ensureModelSelection = () => {
  if (!models.value.length) {
    models.value = [...fallbackModels];
  }

  const exists = models.value.find((m) => m.id === modelName.value);
  if (!exists) {
    modelName.value = models.value[0]?.id || defaultModelId;
  }
};

const applyModelPayload = (payload?: LLMModel[]) => {
  if (payload && payload.length) {
    models.value = normalizeModels(payload);
    modelsSource.value = "api";
  } else {
    models.value = [...fallbackModels];
    modelsSource.value = "fallback";
  }
  ensureModelSelection();
};

const { data: llmModels, execute: fetchModels } = api.listLLMModels();

const loadModels = async () => {
  modelsLoading.value = true;
  await fetchModels();
  modelsLoading.value = false;
  applyModelPayload(llmModels.value?.models);
};

onMounted(loadModels);

watch(
  () => llmModels.value,
  (val) => applyModelPayload(val?.models),
);

// Initialize form with chatbot data if editing
const initializeForm = () => {
  if (props.chatbot && props.mode === "edit") {
    name.value = props.chatbot.name || "";
    description.value = props.chatbot.description || "";
    systemInstructions.value =
      props.chatbot.system_instructions || "You are a helpful AI assistant";
    modelName.value = props.chatbot.model_name || defaultModelId;
    temperatureParam.value = props.chatbot.temperature_param || 0.7;
    maxTokens.value = props.chatbot.max_tokens || 2000;
    useMaxTokens.value =
      (props.chatbot as any).use_max_tokens !== undefined
        ? (props.chatbot as any).use_max_tokens
        : false;
    saveMessages.value =
      props.chatbot.save_messages === undefined
        ? true
        : props.chatbot.save_messages;
    selectedSharedKnowledgeBaseIds.value =
      props.chatbot.shared_knowledge_base_ids || [];
  } else {
    // Reset to defaults for create mode
    name.value = "";
    description.value = "";
    systemInstructions.value = "You are a helpful AI assistant";
    modelName.value = defaultModelId;
    temperatureParam.value = 0.7;
    maxTokens.value = 2000;
    saveMessages.value = true;
    useMaxTokens.value = false;
    selectedSharedKnowledgeBaseIds.value = [];
  }
  ensureModelSelection();
};

// Watch for chatbot changes to reinitialize form
watch(() => props.chatbot, initializeForm, { immediate: true });
watch(() => props.mode, initializeForm);

const handleSubmit = () => {
  if (!name.value.trim()) {
    toast.error("Name is required");
    return;
  }

  const formData: ChatbotFormData = {
    name: name.value,
    description: description.value,
    model_name: modelName.value,
    system_instructions: systemInstructions.value,
    max_tokens: Number(maxTokens.value),
    temperature_param: Number(temperatureParam.value),
    save_messages: Boolean(saveMessages.value),
    use_max_tokens: useMaxTokens.value,
    shared_knowledge_base_ids: [...selectedSharedKnowledgeBaseIds.value],
  };

  emit("submit", formData);
};

const handleInsertPrompt = (prompt: string) => {
  systemInstructions.value = prompt;
  showPromptHelper.value = false;
};

// Computed properties
const submitButtonText = computed(() => {
  return props.mode === "edit" ? "Update Chatbot" : "Create Chatbot";
});

const currentModel = computed(() =>
  models.value.find((m) => m.id === modelName.value),
);
const currentModelHint = computed(
  () => currentModel.value?.hint || formatHint(currentModel.value),
);

// UI helpers
const creativityLabel = computed(() => {
  const t = Number(temperatureParam.value);
  if (t < 0.3) return "Focused";
  if (t < 0.8) return "Balanced";
  return "Creative";
});

const responseLengthLabel = computed(() => {
  const tokens = Number(maxTokens.value);
  if (tokens <= 500) return "Short";
  if (tokens <= 1500) return "Medium";
  if (tokens <= 2500) return "Long";
  return "Very Long";
});
</script>

<template>
  <div class="flex flex-col w-full max-w-3xl">
    <form @submit.prevent="handleSubmit" class="space-y-6">
      <!-- Basic Configuration Card -->
      <ChatSectionCard
        title="Basic Configuration"
        subtitle="Set up your assistant's identity and core behavior"
        color="indigo"
      >
        <template #icon>
          <IconUserCircle class="h-5 w-5" />
        </template>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <Label for="name"
              >Name <span class="text-destructive">*</span></Label
            >
            <Input
              id="name"
              v-model="name"
              placeholder="Customer Support Assistant"
              required
              class="mt-2"
            />
          </div>

          <div>
            <Label for="modelName"
              >AI Model <span class="text-destructive">*</span></Label
            >

            <div class="mt-2">
              <Select v-model="modelName" :disabled="modelsLoading">
                <SelectTrigger id="modelName" class="w-full">
                  <SelectValue placeholder="Select a model" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="m in models" :key="m.id" :value="m.id">
                    <div class="flex flex-col">
                      <span class="font-medium">{{ m.label || m.id }}</span>
                      <span class="text-xs text-muted-foreground">
                        {{ m.hint || formatHint(m) }}
                      </span>
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
              <p class="mt-2 text-xs text-muted-foreground">
                <span>{{ currentModelHint }}</span>
                <span v-if="modelsSource === 'fallback'" class="ml-1">
                  (using fallback list while model catalog loads)
                </span>
              </p>
            </div>
          </div>
        </div>

        <div class="mt-6">
          <Label for="description">Description</Label>
          <Textarea
            id="description"
            v-model="description"
            placeholder="A friendly AI assistant that helps customers with their questions and provides detailed support"
            class="mt-2 min-h-[84px]"
          />
        </div>

        <div class="mt-6">
          <Label for="sharedKnowledgeBases">Shared Knowledge Bases</Label>
          <div class="mt-2">
            <SharedKnowledgeSelect
              id="sharedKnowledgeBases"
              v-model="selectedSharedKnowledgeBaseIds"
              :options="props.sharedKnowledgeBases"
              :placeholder="
                props.sharedKnowledgeBases?.length
                  ? 'Assign knowledge bases...'
                  : 'No shared knowledge bases available'
              "
              :disabled="!props.sharedKnowledgeBases?.length"
            />
          </div>
          <p class="mt-2 text-xs text-muted-foreground">
            <span v-if="props.sharedKnowledgeBases?.length">
              Select any reusable knowledge bases you want this chatbot to
              access.
            </span>
            <span v-else>
              No shared knowledge bases available yet. You can create one from
              the Knowledge Bases section.
            </span>
          </p>
        </div>

        <div class="mt-6">
          <div class="flex items-center justify-between gap-3">
            <Label for="systemInstructions">System Instructions</Label>
            <Button
              variant="outline"
              size="sm"
              type="button"
              class="whitespace-nowrap"
              @click="showPromptHelper = true"
            >
              Prompt Helper
            </Button>
          </div>
          <Textarea
            id="systemInstructions"
            v-model="systemInstructions"
            placeholder="You are a helpful customer support assistant. Be friendly, professional, and provide accurate information. Always ask clarifying questions when needed."
            class="mt-2 min-h-[92px]"
          />
        </div>

        <!-- Submit on small screens inside the first card for convenience -->
        <div class="mt-6">
          <Button
            type="submit"
            :loading="isLoading"
            :disabled="isLoading"
            v-if="mode !== 'create'"
            >{{ submitButtonText }}</Button
          >
        </div>
      </ChatSectionCard>

      <!-- Advanced Settings Card -->
      <ChatSectionCard
        title="Advanced Settings"
        subtitle="Fine‑tune response behavior and performance"
        color="purple"
      >
        <template #icon>
          <IconSun class="h-5 w-5" />
        </template>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
          <!-- Creativity Level -->
          <div>
            <div class="flex items-center justify-between">
              <Label>Creativity Level</Label>
              <div
                class="inline-flex h-7 min-w-7 items-center justify-center rounded-full bg-muted px-2 text-xs font-medium text-muted-foreground"
              >
                {{ temperatureParam.toFixed(1) }}
              </div>
            </div>
            <div class="mt-3">
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                v-model.number="temperatureParam"
                class="w-full appearance-none bg-transparent"
                aria-label="Creativity level"
              />
              <!-- Custom track -->
              <div class="relative mt-2 h-2 rounded-full bg-muted">
                <div
                  class="absolute h-2 rounded-full bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600"
                  :style="{
                    width: `${Number(temperatureParam) * 100}%`,
                  }"
                />
              </div>
              <div
                class="mt-2 flex items-center justify-between text-xs text-muted-foreground"
              >
                <span>Focused</span>
                <span>Balanced</span>
                <span>Creative</span>
              </div>
            </div>
          </div>

          <!-- Response length -->
          <div>
            <div class="flex items-center justify-between">
              <Label>Response Length</Label>
              <div
                class="inline-flex h-7 min-w-7 items-center justify-center rounded-full bg-muted px-2 text-xs font-medium text-muted-foreground"
              >
                {{ maxTokens }}
              </div>
            </div>
            <div class="mt-2 flex items-start justify-between gap-4">
              <p class="text-xs text-muted-foreground max-w-xs">
                When enabled, the model is asked to stop after the
                <span class="font-medium">max tokens</span> value. Setting this
                too low can cause cut off or even empty responses.
              </p>
              <Switch
                :model-value="useMaxTokens"
                @update:model-value="(value) => (useMaxTokens = value)"
              />
            </div>
            <div class="mt-3">
              <input
                type="range"
                min="100"
                max="4000"
                step="100"
                v-model.number="maxTokens"
                :disabled="!useMaxTokens"
                class="w-full appearance-none bg-transparent"
                aria-label="Response length"
              />
              <!-- Custom track -->
              <div class="relative mt-2 h-2 rounded-full bg-muted">
                <div
                  class="absolute h-2 rounded-full bg-gradient-to-r from-green-300 to-yellow-400"
                  :style="{
                    width: `${((Number(maxTokens) - 100) / 3900) * 100}%`,
                  }"
                />
              </div>
              <div
                class="mt-2 flex items-center justify-between text-xs text-muted-foreground"
              >
                <span>Short</span>
                <span>Medium</span>
                <span>Long</span>
                <span>Very Long</span>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-8">
          <div class="flex items-start justify-between gap-4">
            <div>
              <Label>Save Messages</Label>
              <p class="mt-1 text-sm text-muted-foreground">
                Store each conversation in chat history for future reference.
                Turn this off to skip saving conversations.
              </p>
            </div>
            <Switch
              :model-value="saveMessages"
              @update:model-value="(value) => (saveMessages = value)"
            />
          </div>
        </div>

        <div class="mt-8">
          <Button type="submit" :loading="isLoading" :disabled="isLoading">{{
            submitButtonText
          }}</Button>
        </div>
      </ChatSectionCard>
    </form>
    <SystemPromptHelperDialog
      v-model:open="showPromptHelper"
      @insert="handleInsertPrompt"
    />
  </div>
</template>

<style scoped>
/* Improve native range appearance cross‑browser */
input[type="range"] {
  height: 28px;
}
input[type="range"]::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  height: 16px;
  width: 16px;
  border-radius: 9999px;
  background: white;
  border: 2px solid rgb(99 102 241); /* indigo-500 */
  margin-top: -4px; /* center on track */
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}
input[type="range"]::-moz-range-thumb {
  height: 16px;
  width: 16px;
  border: 2px solid rgb(99 102 241);
  border-radius: 9999px;
  background: white;
}
input[type="range"]::-webkit-slider-runnable-track {
  height: 8px;
  background: rgba(0, 0, 0, 0.08);
  border-radius: 9999px;
}
input[type="range"]::-moz-range-track {
  height: 8px;
  background: rgba(0, 0, 0, 0.08);
  border-radius: 9999px;
}
</style>
