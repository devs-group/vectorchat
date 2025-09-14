<template>
  <div></div>
</template>

<script setup lang="ts">
// Redirect to details page as the default tab
import { onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";

const route = useRoute();
const router = useRouter();

onMounted(() => {
  const chatId = route.params.id;
  router.replace(`/chat/${chatId}/details`);
});
</script>
