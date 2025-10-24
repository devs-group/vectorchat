<script setup lang="ts">
import { Teleport } from "vue";
import { Button } from "@/components/ui/button";

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  loginHref: {
    type: String,
    default: "#",
  },
  isChecking: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: "Sign in to test your chatbot",
  },
  description: {
    type: String,
    default:
      "You need to be signed in to interact with the live preview. We'll bring you right back here once you log in.",
  },
  ctaLabel: {
    type: String,
    default: "Sign in",
  },
});
</script>

<template>
  <Teleport to="body">
    <div
      v-if="props.open"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4 backdrop-blur-sm"
    >
      <div
        class="w-full max-w-md rounded-3xl bg-white p-8 text-slate-900 shadow-2xl shadow-blue-500/10"
      >
        <div v-if="props.isChecking" class="space-y-4 text-center">
          <div
            class="mx-auto flex size-12 items-center justify-center rounded-full border-4 border-blue-200 border-t-blue-500 animate-spin"
          ></div>
          <p class="text-sm text-slate-500">Checking your account...</p>
        </div>
        <div v-else class="space-y-6">
          <div class="space-y-2 text-center sm:text-left">
            <h2 class="text-2xl font-semibold text-slate-900">
              {{ props.title }}
            </h2>
            <p class="text-sm text-slate-600">
              {{ props.description }}
            </p>
          </div>
          <div class="flex flex-col gap-3 sm:flex-row sm:justify-end">
            <Button as="a" :href="props.loginHref" class="w-full">
              {{ props.ctaLabel }}
            </Button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
