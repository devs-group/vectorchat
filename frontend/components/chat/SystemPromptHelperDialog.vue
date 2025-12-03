<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { useApiService } from "@/composables/useApiService";
import type { SystemPromptGenerateResponse } from "~/types/api";
import { Loader2 } from "lucide-vue-next";

interface Props {
  open: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "update:open", value: boolean): void;
  (e: "insert", prompt: string): void;
}>();

const apiService = useApiService();
const {
  data: generatedData,
  error: generateError,
  isLoading,
  execute: executeGenerate,
} = apiService.generateSystemPrompt<SystemPromptGenerateResponse>();

const dialogOpen = computed({
  get: () => props.open,
  set: (value: boolean) => emit("update:open", value),
});

const purpose = ref("");
const tone = ref("balanced");
const generatedPrompt = ref("");
const inlineError = ref("");

const stage = computed(() =>
  generatedPrompt.value ? ("result" as const) : ("input" as const),
);

const reset = () => {
  purpose.value = "";
  tone.value = "balanced";
  generatedPrompt.value = "";
  inlineError.value = "";
};

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) {
      reset();
    }
  },
);

const handleGenerate = async () => {
  inlineError.value = "";
  generatedPrompt.value = "";

  if (!purpose.value.trim()) {
    inlineError.value = "Please describe the assistant's purpose first.";
    return;
  }

  await executeGenerate({
    purpose: purpose.value.trim(),
    tone: tone.value,
  });

  if (generateError.value) {
    inlineError.value = "Could not generate a prompt. Please try again.";
    return;
  }

  const response = generatedData.value as SystemPromptGenerateResponse | null;
  if (response?.prompt) {
    generatedPrompt.value = response.prompt;
  } else {
    inlineError.value = "The server returned an empty prompt.";
  }
};

const handleInsert = () => {
  if (!generatedPrompt.value) return;
  emit("insert", generatedPrompt.value);
  reset();
  emit("update:open", false);
};

const handleClose = () => {
  emit("update:open", false);
};
</script>

<template>
  <Dialog v-model:open="dialogOpen">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>System Prompt Helper</DialogTitle>
        <DialogDescription>
          Describe what this assistant should do. We'll draft a concise system
          prompt you can insert directly.
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-5">
        <div v-if="stage === 'input'" class="space-y-4">
          <div class="space-y-2">
            <Label for="prompt-purpose">Assistant purpose</Label>
            <Textarea
              id="prompt-purpose"
              v-model="purpose"
              placeholder="e.g., An assistant that triages customer support tickets and asks clarifying questions."
              class="min-h-[96px]"
            />
          </div>

          <div class="space-y-2">
            <Label for="prompt-tone">Tone</Label>
            <Select v-model="tone">
              <SelectTrigger id="prompt-tone" class="w-full">
                <SelectValue placeholder="Select tone" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="balanced">Balanced</SelectItem>
                <SelectItem value="concise">Concise</SelectItem>
                <SelectItem value="friendly">Friendly</SelectItem>
                <SelectItem value="formal">Formal</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div v-else class="space-y-3">
          <div class="space-y-2">
            <Label for="generated-prompt">Generated prompt</Label>
            <Textarea
              id="generated-prompt"
              v-model="generatedPrompt"
              class="min-h-[180px] font-mono text-sm"
              readonly
            />
          </div>
        </div>

        <p v-if="inlineError" class="text-sm text-destructive">
          {{ inlineError }}
        </p>
      </div>

      <DialogFooter class="mt-2 gap-2">
        <Button variant="ghost" type="button" @click="handleClose">
          Close
        </Button>
        <Button
          v-if="stage === 'input'"
          type="button"
          :disabled="isLoading || !purpose.trim()"
          @click="handleGenerate"
        >
          <span v-if="isLoading" class="inline-flex items-center gap-2">
            <Loader2 class="h-4 w-4 animate-spin" />
            Generating...
          </span>
          <span v-else>
            Generate
          </span>
        </Button>
        <Button
          v-else
          type="button"
          variant="default"
          :disabled="!generatedPrompt"
          @click="handleInsert"
        >
          Insert
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
