<script setup lang="ts">
import { useHead } from "#imports";
import HeroSection from "@/components/HeroSection.vue";
import HowItWorksSection from "@/components/HowItWorksSection.vue";
import WhySection from "@/components/WhySection.vue";

definePageMeta({
  layout: "landing",
});

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
  <div class="flex flex-col">
    <HeroSection />
    <WhySection />
    <HowItWorksSection />
  </div>
</template>
