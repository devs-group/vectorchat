<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>Generate New API Key</DialogTitle>
        <DialogDescription>
          Generate a new API key to access the VectorChat API. Keep your keys
          secure and never share them publicly.
        </DialogDescription>
      </DialogHeader>

      <div class="flex flex-col gap-4 py-4">
        <div class="flex flex-col gap-2">
          <Label for="name">Key Name</Label>
          <Input
            id="name"
            v-model="keyName"
            placeholder="Enter a name for your API key"
          />
        </div>

        <div class="flex flex-col gap-2">
          <Label for="expiresAt">Expires At</Label>
          <DatePicker
            v-model="expiresAtDate"
            :min-date="minDate"
            placeholder="Select expiration date (optional)"
          />
          <p class="text-xs text-muted-foreground">
            Leave empty for no expiration
          </p>
        </div>
      </div>

      <DialogFooter>
        <DialogClose as-child>
          <Button variant="outline" @click="handleCancel">Cancel</Button>
        </DialogClose>
        <Button
          @click="handleGenerate"
          :loading="isLoading"
          :disabled="!keyName.trim()"
        >
          Generate Key
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { CalendarDate, type DateValue } from "@internationalized/date";
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
import { DatePicker } from "@/components/ui/date-picker";

interface Props {
  open: boolean;
  isLoading?: boolean;
}

interface Emits {
  (e: "update:open", value: boolean): void;
  (e: "generate", data: { name: string; expires_at?: string }): void;
}

const props = withDefaults(defineProps<Props>(), {
  isLoading: false,
});

const emit = defineEmits<Emits>();

const keyName = ref("");
const expiresAtDate = ref<DateValue | undefined>();

const isOpen = computed({
  get: () => props.open,
  set: (value) => emit("update:open", value),
});

// Set minimum date to today
const minDate = computed(() => {
  const today = new Date();
  return new CalendarDate(
    today.getFullYear(),
    today.getMonth() + 1,
    today.getDate(),
  );
});

// Reset form when dialog closes
watch(
  () => props.open,
  (newValue) => {
    if (!newValue) {
      keyName.value = "";
      expiresAtDate.value = undefined;
    }
  },
);

const handleCancel = () => {
  isOpen.value = false;
};

const handleGenerate = () => {
  if (!keyName.value.trim()) return;

  const data: { name: string; expires_at?: string } = {
    name: keyName.value.trim(),
  };

  if (expiresAtDate.value) {
    // Convert DateValue to JavaScript Date for ISO string
    const jsDate = new Date(
      expiresAtDate.value.year,
      expiresAtDate.value.month - 1,
      expiresAtDate.value.day,
    );
    data.expires_at = jsDate.toISOString();
  }

  emit("generate", data);
};
</script>
