<template>
  <button
    type="button"
    class="flex-1 inline-flex items-center justify-center gap-2 rounded-full py-2 transition-colors"
    :class="computedClass"
    @click="handleClick"
  >
    <slot name="icon" />
    <slot />
  </button>
</template>

<script setup lang="ts">
import { computed, inject } from 'vue'

interface Props {
  value: string | number
  disabled?: boolean
  class?: string
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  class: ''
})

// Get context from parent PillTabs component
const tabsContext = inject<{
  activeTab: { value: string | number | undefined }
  setActiveTab: (value: string | number) => void
}>('pill-tabs')

if (!tabsContext) {
  console.warn('PillTab must be used within PillTabs component')
}

const isActive = computed(() => {
  return tabsContext?.activeTab.value === props.value
})

const computedClass = computed(() => {
  const baseClasses = []

  if (isActive.value) {
    baseClasses.push('bg-background shadow-sm')
  } else {
    baseClasses.push('text-muted-foreground hover:text-foreground')
  }

  if (props.disabled) {
    baseClasses.push('opacity-50 cursor-not-allowed')
  }

  if (props.class) {
    baseClasses.push(props.class)
  }

  return baseClasses.join(' ')
})

const handleClick = () => {
  if (!props.disabled && tabsContext) {
    tabsContext.setActiveTab(props.value)
  }
}
</script>
