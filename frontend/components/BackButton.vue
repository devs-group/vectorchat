<template>
  <button
    type="button"
    class="inline-flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
    @click="handleBack"
  >
    <IconArrowLeft class="h-4 w-4" />
    Back
  </button>
</template>

<script setup lang="ts">
import IconArrowLeft from "@/components/icons/IconArrowLeft.vue";

const props = withDefaults(defineProps<{ fallback?: string; to?: string }>(), {
  fallback: "/",
});

const router = useRouter();

const handleBack = () => {
  if (props.to) {
    router.push(props.to);
    return;
  }

  if (process.client && window.history.length > 1) {
    router.back();
    return;
  }

  router.push(props.fallback);
};
</script>
