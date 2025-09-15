<script setup lang="ts">
import { Slot } from 'reka-ui'
</script>

<template>
  <p data-slot="drawer-description" class="text-sm text-muted-foreground">
    <Slot />
  </p>
</template>

