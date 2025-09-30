<template>
  <div class="flex min-h-[calc(100vh-3.5rem)] items-center justify-center">
    <div class="mx-auto w-full max-w-lg space-y-6 text-center">
      <div class="space-y-2">
        <h1 class="text-3xl font-semibold tracking-tight">
          Something went wrong
        </h1>
        <p class="text-muted-foreground">
          {{ displayMessage }}
        </p>
      </div>

      <div
        v-if="errorDetails.length"
        class="rounded-md border border-border bg-muted/40 p-4 text-left text-sm"
      >
        <p class="font-medium text-foreground">Additional details</p>
        <ul class="mt-2 space-y-1 text-muted-foreground">
          <li v-for="(detail, index) in errorDetails" :key="index">
            • {{ detail }}
          </li>
        </ul>
      </div>

      <div class="flex flex-wrap justify-center gap-3 pt-2">
        <Button variant="outline" @click="handleRetry" :disabled="isLoading">
          Try again
        </Button>
        <Button @click="handleGoHome" :disabled="isLoading">Go to login</Button>
      </div>

      <p v-if="isLoading" class="text-xs text-muted-foreground">
        Loading error details…
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Button } from "~/components/ui/button";

type KratosErrorDetails = {
  message?: string;
  reason?: string;
};

type KratosErrorResponse = {
  error?: {
    code?: number;
    status?: string;
    message?: string;
    reason?: string;
    details?: KratosErrorDetails[];
  };
};

const config = useRuntimeConfig();
const route = useRoute();
const router = useRouter();

const isLoading = ref(false);
const errorResponse = ref<KratosErrorResponse | null>(null);
const fallbackMessage = "We could not complete your request. Please try again.";

const displayMessage = computed(() => {
  const data = errorResponse.value?.error;
  return data?.reason || data?.message || fallbackMessage;
});

const errorDetails = computed(() => {
  const details = errorResponse.value?.error?.details || [];
  return details
    .map((detail) => detail.reason || detail.message)
    .filter((msg): msg is string => Boolean(msg));
});

const loadError = async () => {
  const id = typeof route.query.id === "string" ? route.query.id : null;
  if (!id) {
    errorResponse.value = null;
    return;
  }

  try {
    isLoading.value = true;
    errorResponse.value = await $fetch<KratosErrorResponse>(
      `${config.public.kratosPublicUrl}/self-service/errors`,
      {
        params: { id },
        credentials: "include",
      },
    );
  } catch (error) {
    console.error("Failed to load Kratos error flow", error);
    errorResponse.value = null;
  } finally {
    isLoading.value = false;
  }
};

const handleRetry = () => {
  const returnTo = route.query.return_to;
  if (typeof returnTo === "string" && returnTo) {
    router.push(returnTo);
  } else {
    router.push({ path: "/login" });
  }
};

const handleGoHome = () => {
  router.push({ path: "/login" });
};

onMounted(() => {
  loadError();
});

watch(
  () => route.query.id,
  (newId, oldId) => {
    if (newId !== oldId) {
      loadError();
    }
  },
);

useHead({ title: "Error" });
</script>
