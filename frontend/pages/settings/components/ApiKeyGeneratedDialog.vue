<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>API Key Generated</DialogTitle>
        <DialogDescription>
          Your API key has been generated successfully. Copy it now as you won't
          be able to see it again.
        </DialogDescription>
      </DialogHeader>

      <div class="flex flex-col gap-4 py-4">
        <div v-if="apiKey" class="flex flex-col gap-2">
          <Label>API Key</Label>
          <code class="p-2 bg-gray-100 rounded text-sm font-mono break-all">
            {{ apiKey }}
          </code>
        </div>

        <div
          class="flex items-center justify-between p-3 bg-amber-50 border border-amber-200 rounded-md"
        >
          <div class="flex items-center gap-2">
            <IconAlertCircle class="h-4 w-4 text-amber-600" />
            <span class="text-sm text-amber-800">
              Store this key securely - it won't be shown again
            </span>
          </div>
        </div>
      </div>

      <DialogFooter>
        <Button
          @click="handleCopy"
          :disabled="copyState === 'copied'"
          class="flex-1"
        >
          <IconCopy class="mr-2 h-4 w-4" />
          {{ copyState === "copied" ? "Copied!" : "Copy API Key" }}
        </Button>
        <DialogClose as-child>
          <Button variant="outline" @click="handleClose"> Close </Button>
        </DialogClose>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogClose,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import IconAlertCircle from "@/components/icons/IconAlertCircle.vue";
import IconCopy from "@/components/icons/IconCopy.vue";

interface Props {
  open: boolean;
  apiKey?: string;
}

interface Emits {
  (e: "update:open", value: boolean): void;
  (e: "close"): void;
}

const props = withDefaults(defineProps<Props>(), {
  apiKey: "",
});

const emit = defineEmits<Emits>();

const isVisible = ref(true);
const copyState = ref<"idle" | "copied">("idle");
const keyInput = ref<HTMLInputElement>();

const isOpen = computed({
  get: () => props.open,
  set: (value) => emit("update:open", value),
});

// Reset state when dialog opens
watch(
  () => props.open,
  (newValue) => {
    if (newValue) {
      isVisible.value = true;
      copyState.value = "idle";
    }
  },
);

const toggleVisibility = () => {
  isVisible.value = !isVisible.value;
};

const handleCopy = async () => {
  if (!props.apiKey) return;

  try {
    await navigator.clipboard.writeText(props.apiKey);
    copyState.value = "copied";

    // Reset copy state after 2 seconds
    setTimeout(() => {
      copyState.value = "idle";
    }, 2000);
  } catch (err) {
    console.error("Failed to copy API key:", err);
  }
};

const handleClose = () => {
  isOpen.value = false;
  emit("close");
};
</script>
