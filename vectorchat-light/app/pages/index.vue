<script setup lang="ts">
import { useHead } from "#imports";
import LandingMenu from "@/components/LandingMenu.vue";
import HeroSection from "@/components/HeroSection.vue";
import HowItWorksSection from "@/components/HowItWorksSection.vue";
import WhySection from "@/components/WhySection.vue";

useHead({
  title: "VectorChat Light | Free Chatbot in Under 2 Minutes",
  meta: [
    {
      name: "description",
      content:
        "Type in your website address and launch an AI chatbot in minutes. VectorChat Light keeps visitors engaged without extra engineering.",
    },
    {
      property: "og:title",
      content: "VectorChat Light | Free Chatbot in Under 2 Minutes",
    },
    {
      property: "og:description",
      content:
        "Most sites are static walls of text. Let visitors interact, get instant answers, and stay engaged with VectorChat Light.",
    },
  ],
});
</script>

<template>
  <div class="relative min-h-screen bg-background text-foreground">
    <LandingMenu />

    <main class="flex flex-col">
      <HeroSection />
      <WhySection />
      <HowItWorksSection />
    </main>
  </div>
</template>
