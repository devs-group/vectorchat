<template>
  <div class="flex rounded-full bg-muted/60 p-1 text-sm" :class="containerClass">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { provide, ref, type Ref } from 'vue'

interface Props {
  modelValue?: string | number
  containerClass?: string
}

interface Emits {
  (e: 'update:modelValue', value: string | number): void
}

const props = withDefaults(defineProps<Props>(), {
  containerClass: ''
})

const emit = defineEmits<Emits>()

const activeTab = ref(props.modelValue) as Ref<string | number | undefined>

// Watch for external changes to modelValue
watch(() => props.modelValue, (newValue) => {
  activeTab.value = newValue
})

// Provide context to child components
provide('pill-tabs', {
  activeTab,
  setActiveTab: (value: string | number) => {
    activeTab.value = value
    emit('update:modelValue', value)
  }
})
</script>
