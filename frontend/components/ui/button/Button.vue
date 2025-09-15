<script setup lang="ts">
import { computed } from 'vue'
import IconSpinner from '@/components/icons/IconSpinner.vue'

const props = withDefaults(defineProps<{
  variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link'
  size?: 'default' | 'sm' | 'lg' | 'icon'
  loading?: boolean
}>(), {
  variant: 'default',
  size: 'default',
  loading: false
})

const classes = computed(() => {
  return [
    'inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
    {
      'bg-primary text-primary-foreground hover:bg-primary/90': props.variant === 'default',
      'bg-destructive text-destructive-foreground hover:bg-destructive/90': props.variant === 'destructive',
      'border border-input bg-background hover:bg-accent hover:text-accent-foreground': props.variant === 'outline',
      'bg-secondary text-secondary-foreground hover:bg-secondary/80': props.variant === 'secondary',
      'hover:bg-accent hover:text-accent-foreground': props.variant === 'ghost',
      'text-primary underline-offset-4 hover:underline': props.variant === 'link',
      'h-10 px-4 py-2': props.size === 'default',
      'h-9 rounded-md px-3': props.size === 'sm',
      'h-11 rounded-md px-8': props.size === 'lg',
      'h-10 w-10': props.size === 'icon',
    }
  ]
})
</script>

<template>
  <button
    :class="classes"
    :disabled="loading"
    v-bind="$attrs"
  >
    <IconSpinner v-if="loading" class="mr-2 h-4 w-4 animate-spin" />
    <slot />
  </button>
</template>
