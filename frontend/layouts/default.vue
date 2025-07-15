<template>
  <div class="bg-background">
    <header
      class="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60"
    ></header>
    <main class="container py-6 mx-auto">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
// Layout logic can be added here if needed
</script>
