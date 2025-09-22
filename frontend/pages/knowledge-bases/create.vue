<script setup lang="ts">
definePageMeta({
  layout: "authenticated",
});

import { reactive } from "vue";
import { useRouter } from "vue-router";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useApiService } from "@/composables/useApiService";
import BackButton from "@/components/BackButton.vue";
import IconUserCircle from "@/components/icons/IconUserCircle.vue";

const router = useRouter();
const apiService = useApiService();

const {
  execute: createKnowledgeBase,
  isLoading,
  error,
  data,
} = apiService.createSharedKnowledgeBase();

const form = reactive({
  name: "",
  description: "",
});

const handleSubmit = async () => {
  if (!form.name.trim()) {
    return;
  }

  await createKnowledgeBase({
    name: form.name.trim(),
    description: form.description.trim() || undefined,
  });

  if (!error.value) {
    const kb = data.value;
    if (kb) {
      router.push(`/knowledge-bases/${kb.id}`);
    } else {
      router.push("/knowledge-bases");
    }
  }
};

const cancel = () => {
  router.push("/knowledge-bases");
};
</script>

<template>
  <div class="max-w-3xl space-y-6">
    <BackButton fallback="/knowledge-bases" />
    <div>
      <h1 class="text-3xl font-bold tracking-tight">New Knowledge Base</h1>
      <p class="text-sm text-muted-foreground">
        Create a shared knowledge base that can be linked to multiple chatbots.
      </p>
    </div>

    <ChatSectionCard
      title="Basic Configuration"
      subtitle="Describe your shared knowledge base"
      color="green"
    >
      <template #icon>
        <IconUserCircle class="h-5 w-5" />
      </template>
      <form @submit.prevent="handleSubmit">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div>
            <Label for="kb-name">Name</Label>
            <Input
              id="kb-name"
              v-model="form.name"
              placeholder="Support Handbook"
              required
              class="mt-2"
            />
          </div>

          <div>
            <Label for="kb-description">Description</Label>
            <Textarea
              id="kb-description"
              v-model="form.description"
              placeholder="Short summary of what this knowledge base covers"
              class="mt-2 min-h-[96px]"
            />
          </div>
        </div>

        <div class="flex items-center gap-3 mt-5">
          <Button type="button" variant="outline" @click="cancel"
            >Cancel</Button
          >
          <Button type="submit" :loading="isLoading" :disabled="isLoading">
            Create
          </Button>
        </div>
      </form>
    </ChatSectionCard>
  </div>
</template>
