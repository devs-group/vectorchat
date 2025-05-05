<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="fixed inset-0 z-50 bg-black/80 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
      @click="$emit('update:modelValue', false)"
    ></div>
    <div
      v-if="modelValue"
      class="fixed inset-0 z-50 flex items-center justify-center"
      @click="$emit('update:modelValue', false)"
    >
      <slot />
    </div>
  </Teleport>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: boolean
}>()

defineEmits<{
  'update:modelValue': [value: boolean]
}>()
</script> 