<script setup lang="ts">
import { useRouter, useRoute } from "vue-router";
import { onMounted } from "vue";

definePageMeta({
  layout: "authenticated",
});

const router = useRouter();
const route = useRoute();

onMounted(() => {
  const id = route.params.id;
  router.replace(`/knowledge-bases/${id}/details`);
});
</script>

<template>
  <div class="flex min-h-[200px] items-center justify-center">
    <div
      class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary"
    ></div>
  </div>
</template>
