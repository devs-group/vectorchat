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
        <template v-if="credentials">
          <div class="flex flex-col gap-2">
            <Label>Client ID</Label>
            <code class="p-2 bg-gray-100 rounded text-sm font-mono break-all">
              {{ credentials.clientId }}
            </code>
          </div>

          <div class="flex flex-col gap-2">
            <div class="flex items-center justify-between">
              <Label>Client Secret</Label>
              <button
                type="button"
                class="text-xs text-primary underline"
                @click="toggleVisibility"
              >
                {{ isVisible ? "Hide" : "Show" }}
              </button>
            </div>
            <code class="p-2 bg-gray-100 rounded text-sm font-mono break-all">
              {{ isVisible ? credentials.clientSecret : maskedSecret }}
            </code>
          </div>

          <div v-if="credentials.expiresAt" class="text-sm text-muted-foreground">
            Expires on {{ formatDate(credentials.expiresAt) }}
          </div>

          <div
            class="flex items-center justify-between p-3 bg-amber-50 border border-amber-200 rounded-md"
          >
            <div class="flex items-center gap-2">
              <IconAlertCircle class="h-4 w-4 text-amber-600" />
              <span class="text-sm text-amber-800">
                Store these credentials securely - the secret will not be shown
                again.
              </span>
            </div>
          </div>
        </template>
      </div>

      <DialogFooter>
        <div class="flex flex-col gap-2 flex-1">
          <div class="flex gap-2">
            <Button
              @click="handleCopySecret"
              :disabled="copyState.secret === 'copied'"
              class="flex-1"
            >
              <IconCopy class="mr-2 h-4 w-4" />
              {{ copyState.secret === "copied" ? "Secret copied" : "Copy Secret" }}
            </Button>
            <Button
              variant="secondary"
              @click="handleCopyId"
              :disabled="copyState.id === 'copied'"
            >
              <IconCopy class="mr-2 h-4 w-4" />
              {{ copyState.id === "copied" ? "ID copied" : "Copy ID" }}
            </Button>
          </div>
          <DialogClose as-child>
            <Button variant="outline" @click="handleClose"> Close </Button>
          </DialogClose>
        </div>
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
import { Label } from "@/components/ui/label";
import IconAlertCircle from "@/components/icons/IconAlertCircle.vue";
import IconCopy from "@/components/icons/IconCopy.vue";

interface Props {
  open: boolean;
  credentials?: {
    clientId: string;
    clientSecret: string;
    name?: string | null;
    expiresAt?: string | null;
  } | null;
}

interface Emits {
  (e: "update:open", value: boolean): void;
  (e: "close"): void;
}

const props = withDefaults(defineProps<Props>(), {
  credentials: null,
});

const emit = defineEmits<Emits>();

const isVisible = ref(true);
const copyState = ref({ secret: "idle", id: "idle" } as {
  secret: "idle" | "copied";
  id: "idle" | "copied";
});

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
      copyState.value = { secret: "idle", id: "idle" };
    }
  },
);

const toggleVisibility = () => {
  isVisible.value = !isVisible.value;
};

const maskedSecret = computed(() => {
  const length = props.credentials?.clientSecret.length ?? 16;
  return "•".repeat(Math.min(Math.max(length, 16), 48));
});

const formatDate = (value: string) => {
  const date = new Date(value);
  return date.toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

const handleCopySecret = async () => {
  if (!props.credentials?.clientSecret) return;

  try {
    await navigator.clipboard.writeText(props.credentials.clientSecret);
    copyState.value.secret = "copied";

    // Reset copy state after 2 seconds
    setTimeout(() => {
      copyState.value.secret = "idle";
    }, 2000);
  } catch (err) {
    console.error("Failed to copy client secret:", err);
  }
};

const handleCopyId = async () => {
  if (!props.credentials?.clientId) return;

  try {
    await navigator.clipboard.writeText(props.credentials.clientId);
    copyState.value.id = "copied";
    setTimeout(() => {
      copyState.value.id = "idle";
    }, 2000);
  } catch (err) {
    console.error("Failed to copy client ID:", err);
  }
};

const handleClose = () => {
  isOpen.value = false;
  emit("close");
};
</script>
