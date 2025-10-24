<script setup lang="ts">
import LandingMenu from "@/components/LandingMenu.vue";
import FooterSection from "@/components/FooterSection.vue";
</script>

<template>
  <div class="relative min-h-screen bg-background text-foreground">
    <LandingMenu />

    <main>
      <slot />
    </main>

    <FooterSection />
  </div>
</template>
