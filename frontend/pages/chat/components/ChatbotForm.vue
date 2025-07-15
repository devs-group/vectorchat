<script setup lang="ts">
import { toast } from "vue-sonner";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import type { ChatbotResponse } from "~/types/api";

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
  } else {
    // Reset to defaults for create mode
    name.value = "";
    description.value = "";
    systemInstructions.value = "You are a helpful AI assistant";
    modelName.value = "gpt-4";
    temperatureParam.value = 0.7;
    maxTokens.value = 2000;
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

const formTitle = computed(() => {
  return props.mode === "edit" ? "Edit Chatbot" : "Create a New Chatbot";
});
</script>

<template>
  <div class="flex flex-col w-full max-w-xl justify-start">
    <h1 class="mb-8 text-2xl font-bold tracking-tight text-left">
      {{ formTitle }}
    </h1>
    <form @submit.prevent="handleSubmit" class="space-y-6 w-full">
      <div>
        <Label for="name">Name <span class="text-destructive">*</span></Label>
        <Input
          id="name"
          v-model="name"
          placeholder="My AI Assistant"
          required
          class="mt-2"
        />
      </div>

      <div>
        <Label for="description">Description</Label>
        <Textarea
          id="description"
          v-model="description"
          placeholder="A helpful AI assistant for my project"
          class="mt-2 min-h-[80px]"
        />
      </div>

      <div>
        <Label for="systemInstructions">System Instructions</Label>
        <Textarea
          id="systemInstructions"
          v-model="systemInstructions"
          placeholder="You are a helpful AI assistant"
          class="mt-2 min-h-[80px]"
        />
      </div>

      <div>
        <Label for="modelName">Model Name</Label>
        <Input
          id="modelName"
          v-model="modelName"
          placeholder="gpt-4"
          class="mt-2"
        />
      </div>

      <div class="flex gap-4">
        <div class="flex-1">
          <Label for="temperatureParam">Temperature</Label>
          <Input
            id="temperatureParam"
            v-model="temperatureParam"
            type="number"
            step="0.01"
            min="0"
            max="2"
            placeholder="0.7"
            class="mt-2"
          />
        </div>
        <div class="flex-1">
          <Label for="maxTokens">Max Tokens</Label>
          <Input
            id="maxTokens"
            v-model="maxTokens"
            type="number"
            min="1"
            max="4000"
            placeholder="2000"
            class="mt-2"
          />
        </div>
      </div>

      <Button
        type="submit"
        :loading="isLoading"
        :disabled="isLoading"
        class="mt-4 w-full sm:w-auto"
      >
        {{ submitButtonText }}
      </Button>
    </form>
  </div>
</template>
