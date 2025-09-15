<script setup lang="ts">
import { Slot } from 'reka-ui'
</script>

<template>
  <h2 data-slot="drawer-title" class="text-lg font-semibold leading-none tracking-tight">
    <Slot />
  </h2>
</template>

