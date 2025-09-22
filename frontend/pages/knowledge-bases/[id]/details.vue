<script setup lang="ts">
definePageMeta({
  layout: "authenticated",
});

import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import KnowledgeBase from "@/pages/chat/[id]/components/KnowledgeBase.vue";
import { useApiService } from "@/composables/useApiService";
import BackButton from "@/components/BackButton.vue";
import type {
  SharedKnowledgeBase,
  SharedKnowledgeBaseUpdateRequest,
} from "~/types/api";
import IconUserCircle from "@/components/icons/IconUserCircle.vue";

const route = useRoute();
const router = useRouter();
const kbId = computed(() => route.params.id as string);
const apiService = useApiService();

const knowledgeBase = ref<SharedKnowledgeBase | null>(null);
const form = reactive({ name: "", description: "" });

const {
  execute: fetchKnowledgeBase,
  data,
  isLoading,
} = apiService.getSharedKnowledgeBase();

const {
  execute: updateKnowledgeBase,
  isLoading: isUpdating,
  error: updateError,
} = apiService.updateSharedKnowledgeBase();

onMounted(async () => {
  if (!kbId.value) return;
  await fetchKnowledgeBase(kbId.value);
  if (data.value) {
    knowledgeBase.value = data.value;
    form.name = data.value.name;
    form.description = data.value.description ?? "";
  }
});

watch(kbId, async (newId) => {
  if (!newId) return;
  await fetchKnowledgeBase(newId as string);
  if (data.value) {
    knowledgeBase.value = data.value;
    form.name = data.value.name;
    form.description = data.value.description ?? "";
  }
});

const handleUpdate = async () => {
  if (!knowledgeBase.value) return;
  const body: SharedKnowledgeBaseUpdateRequest = {
    name: form.name.trim() || undefined,
    description: form.description.trim() || undefined,
  };

  await updateKnowledgeBase({ id: knowledgeBase.value.id, body });
  if (!updateError.value && knowledgeBase.value) {
    knowledgeBase.value = {
      ...knowledgeBase.value,
      name: form.name,
      description: form.description,
      updated_at: new Date().toISOString(),
    };
  }
};
</script>

<template>
  <div class="max-w-3xl space-y-8">
    <BackButton fallback="/knowledge-bases" />
    <div v-if="isLoading" class="flex justify-center py-12">
      <div
        class="h-10 w-10 animate-spin rounded-full border-b-2 border-primary"
      ></div>
    </div>

    <template v-else-if="knowledgeBase">
      <ChatSectionCard
        title="Basic Configuration"
        subtitle="Describe your shared knowledge base"
        color="green"
      >
        <template #icon>
          <IconUserCircle class="h-5 w-5" />
        </template>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div>
            <Label for="kb-name">Name</Label>
            <Input
              id="kb-name"
              v-model="form.name"
              class="mt-2"
              placeholder="Support Knowledge Base"
            />
          </div>
          <div>
            <Label for="kb-description">Description</Label>
            <Textarea
              id="kb-description"
              v-model="form.description"
              class="mt-2 min-h-[96px]"
              placeholder="Optional summary of this knowledge base"
            />
          </div>
          <div>
            <Button
              type="button"
              :loading="isUpdating"
              :disabled="isUpdating"
              @click="handleUpdate"
            >
              Save Changes
            </Button>
          </div>
        </div>
      </ChatSectionCard>

      <div class="border-t border-border pt-8">
        <KnowledgeBase :resource-id="kbId" scope="shared" />
      </div>
    </template>

    <div v-else class="rounded-xl border border-border bg-card p-6 text-center">
      <p class="text-muted-foreground">Knowledge base not found.</p>
      <Button
        class="mt-4"
        variant="outline"
        @click="router.push('/knowledge-bases')"
      >
        Back to Knowledge Bases
      </Button>
    </div>
  </div>
</template>
