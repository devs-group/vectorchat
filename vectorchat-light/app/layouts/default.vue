<script setup lang="ts">
import LandingMenu from "@/components/LandingMenu.vue";
</script>

<template>
  <div
    class="relative min-h-screen bg-gradient-to-br from-slate-50 to-blue-50 text-foreground"
  >
    <LandingMenu />

    <main>
      <slot />
    </main>
  </div>
</template>
