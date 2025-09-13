<script lang="ts" setup>
import { Toaster } from "@/components/ui/sonner";
import "vue-sonner/style.css";
</script>
<template>
  <div>
    <Toaster closeButton />

    <NuxtLayout>
      <NuxtPage></NuxtPage>
    </NuxtLayout>
  </div>
</template>
