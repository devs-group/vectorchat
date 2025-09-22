<script setup lang="ts">
import { computed, nextTick, ref } from "vue";
import {
  TagsInput,
  TagsInputInput,
  TagsInputItem,
  TagsInputItemDelete,
} from "@/components/ui/tags-input";
import {
  Combobox,
  ComboboxAnchor,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxItemIndicator,
  ComboboxList,
  ComboboxSeparator,
  ComboboxViewport,
} from "@/components/ui/combobox";
import type { SharedKnowledgeBase } from "~/types/api";
import { Check } from "lucide-vue-next";

interface Props {
  modelValue: string[];
  options: SharedKnowledgeBase[];
  placeholder?: string;
  disabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: () => [],
  options: () => [],
  placeholder: "Assign knowledge bases...",
  disabled: false,
});

const emit = defineEmits<{
  (event: "update:modelValue", value: string[]): void;
}>();

const updateValue = (value: string[]) => {
  emit("update:modelValue", value);
};

const comboboxOpen = ref(false);
const comboboxInputRef = ref<InstanceType<typeof ComboboxInput> | null>(null);

const optionLookup = computed(() => {
  const map = new Map<string, SharedKnowledgeBase>();
  for (const option of props.options ?? []) {
    if (option?.id) {
      map.set(option.id, option);
    }
  }
  return map;
});

const selectedValue = computed({
  get: () => props.modelValue ?? [],
  set: (value: string[]) => updateValue(value),
});

const disabledState = computed(() => props.disabled || !props.options?.length);

const focusComboboxSearch = async () => {
  await nextTick();
  const el = comboboxInputRef.value?.$el?.querySelector(
    "input",
  ) as HTMLInputElement | null;
  if (el) {
    el.focus();
    el.select();
  }
};

const handleFocusAnchor = () => {
  if (disabledState.value) return;
  comboboxOpen.value = true;
  focusComboboxSearch();
};

const handleInputKeydown = (event: KeyboardEvent) => {
  if (event.key === "ArrowDown" || event.key === "Enter") {
    event.preventDefault();
    comboboxOpen.value = true;
    focusComboboxSearch();
  }
};
</script>

<template>
  <Combobox
    v-model="selectedValue"
    v-model:open="comboboxOpen"
    multiple
    :disabled="disabledState"
    :open-on-focus="true"
    :open-on-click="true"
    class="w-full"
  >
    <ComboboxAnchor as-child class="w-full">
      <div class="w-full" @click="handleFocusAnchor">
        <TagsInput
          v-model="selectedValue"
          :disabled="disabledState"
          :add-on-blur="false"
          :add-on-tab="false"
          :add-on-paste="false"
          class="w-full"
        >
          <TagsInputItem
            v-for="value in selectedValue"
            :key="value"
            :value="value"
            class="h-7 bg-secondary text-xs"
          >
            <span class="pl-2 pr-1 text-xs font-medium">
              {{ optionLookup.get(value)?.name ?? "Unknown" }}
            </span>
            <TagsInputItemDelete @click.stop />
          </TagsInputItem>

          <TagsInputInput
            :disabled="disabledState"
            :placeholder="selectedValue.length ? '' : placeholder"
            readonly
            class="min-h-7 flex-1 border-none px-0 py-1 text-sm"
            @focus="handleFocusAnchor"
            @keydown="handleInputKeydown"
          />
        </TagsInput>
      </div>
    </ComboboxAnchor>

    <ComboboxList class="min-w-[260px] max-w-sm">
      <ComboboxSeparator class="bg-border" />
      <ComboboxViewport class="max-h-60 overflow-y-auto p-1">
        <ComboboxEmpty
          class="px-2 py-6 text-center text-sm text-muted-foreground"
        >
          No knowledge bases found
        </ComboboxEmpty>
        <ComboboxItem
          v-for="option in props.options"
          :key="option.id"
          :value="option.id"
          :text-value="option.name"
          class="flex items-start gap-2 rounded-md px-2 py-1.5"
        >
          <div class="flex-1">
            <div class="text-sm font-medium text-foreground">
              {{ option.name }}
            </div>
            <div
              v-if="option.description"
              class="text-xs text-muted-foreground"
            >
              {{ option.description }}
            </div>
          </div>
          <ComboboxItemIndicator>
            <Check class="h-4 w-4 text-primary" />
          </ComboboxItemIndicator>
        </ComboboxItem>
      </ComboboxViewport>
    </ComboboxList>
  </Combobox>
</template>
