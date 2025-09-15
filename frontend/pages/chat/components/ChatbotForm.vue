<script setup lang="ts">
import { toast } from "vue-sonner";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import type { ChatbotResponse } from "~/types/api";
import IconUserCircle from "@/components/icons/IconUserCircle.vue";
import IconSun from "@/components/icons/IconSun.vue";

interface Props {
  chatbot?: ChatbotResponse | null;
  isLoading?: boolean;
  mode?: "create" | "edit";
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
}

const props = withDefaults(defineProps<Props>(), {
  chatbot: null,
  isLoading: false,
  mode: "create",
});

const emit = defineEmits<EmitEvents>();

// Form reactive data
const name = ref("");
const description = ref("");
const systemInstructions = ref("You are a helpful AI assistant");
const modelName = ref("gpt-4");
const temperatureParam = ref(0.7);
const maxTokens = ref(2000);
const saveMessages = ref(true);

// Initialize form with chatbot data if editing
const initializeForm = () => {
  if (props.chatbot && props.mode === "edit") {
    name.value = props.chatbot.name || "";
    description.value = props.chatbot.description || "";
    systemInstructions.value =
      props.chatbot.system_instructions || "You are a helpful AI assistant";
    modelName.value = props.chatbot.model_name || "gpt-4";
    temperatureParam.value = props.chatbot.temperature_param || 0.7;
    maxTokens.value = props.chatbot.max_tokens || 2000;
    saveMessages.value =
      props.chatbot.save_messages === undefined
        ? true
        : props.chatbot.save_messages;
  } else {
    // Reset to defaults for create mode
    name.value = "";
    description.value = "";
    systemInstructions.value = "You are a helpful AI assistant";
    modelName.value = "gpt-4";
    temperatureParam.value = 0.7;
    maxTokens.value = 2000;
    saveMessages.value = true;
  }
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
  };

  emit("submit", formData);
};

// Computed properties
const submitButtonText = computed(() => {
  if (props.isLoading) {
    return props.mode === "edit" ? "Updating..." : "Creating...";
  }
  return props.mode === "edit" ? "Update Chatbot" : "Create Chatbot";
});

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

const models = [
  {
    id: "gpt-4",
    name: "GPT-4",
    hint: "Most capable, best for complex tasks",
  },
  { id: "gpt-4o", name: "GPT-4o", hint: "Fast multimodal reasoning" },
  { id: "gpt-3.5-turbo", name: "GPT-3.5 Turbo", hint: "Great speed & value" },
];
</script>

<template>
  <div class="flex flex-col w-full max-w-3xl">
    <form @submit.prevent="handleSubmit" class="space-y-6">
      <!-- Basic Configuration Card -->
      <div class="rounded-2xl border border-border bg-card shadow-sm">
        <div class="px-6 py-5 border-b border-border/70">
          <div class="flex items-start gap-3">
            <div
              class="mt-0.5 inline-flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-blue-500 to-indigo-500 text-white shadow-sm"
            >
              <IconUserCircle class="h-5 w-5" />
            </div>
            <div class="">
              <h2 class="text-lg font-medium">Basic Configuration</h2>
              <p class="text-sm text-muted-foreground">
                Set up your assistant's identity and core behavior
              </p>
            </div>
          </div>
        </div>

        <div class="p-6 md:p-8">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <Label for="name"
                >Assistant Name <span class="text-destructive">*</span></Label
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
              <Label for="modelName">AI Model</Label>
              <div class="mt-2">
                <select
                  id="modelName"
                  v-model="modelName"
                  class="block w-full h-9 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/70 focus-visible:ring-offset-2"
                >
                  <option v-for="m in models" :key="m.id" :value="m.id">
                    {{ m.name }}
                  </option>
                </select>
                <p class="mt-2 text-xs text-muted-foreground">
                  {{ models.find((m) => m.id === modelName)?.hint }}
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
            <Label for="systemInstructions">System Instructions</Label>
            <Textarea
              id="systemInstructions"
              v-model="systemInstructions"
              placeholder="You are a helpful customer support assistant. Be friendly, professional, and provide accurate information. Always ask clarifying questions when needed."
              class="mt-2 min-h-[92px]"
            />
          </div>

          <!-- Submit on small screens inside the first card for convenience -->
          <div class="mt-6 md:hidden">
            <Button
              type="submit"
              :loading="isLoading"
              :disabled="isLoading"
              class="w-full"
              >{{ submitButtonText }}</Button
            >
          </div>
        </div>
      </div>

      <!-- Advanced Settings Card -->
      <div class="rounded-2xl border border-border bg-card shadow-sm">
        <div class="px-6 py-5 border-b border-border/70">
          <div class="flex items-start gap-3">
            <div
              class="mt-0.5 inline-flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-fuchsia-500 to-violet-500 text-white shadow-sm"
            >
              <IconSun class="h-5 w-5" />
            </div>
            <div>
              <h2 class="text-lg font-medium">Advanced Settings</h2>
              <p class="text-sm text-muted-foreground">
                Fine‑tune response behavior and performance
              </p>
            </div>
          </div>
        </div>

        <div class="p-6 md:p-8">
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
              <div class="mt-3">
                <input
                  type="range"
                  min="100"
                  max="4000"
                  step="100"
                  v-model.number="maxTokens"
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
        </div>
      </div>
    </form>
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
