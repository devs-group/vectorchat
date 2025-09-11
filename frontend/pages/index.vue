<template>
  <div></div>
</template>

<script setup lang="ts">
// Redirect to /chat when accessing the root path
await navigateTo('/chat', { replace: true })
</script>
