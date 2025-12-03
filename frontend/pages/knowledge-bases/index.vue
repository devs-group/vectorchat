<script setup lang="ts">
definePageMeta({
  layout: "authenticated",
});

import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import AppResourceCard from "@/components/AppResourceCard.vue";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import IconArrowLeftLong from "@/components/icons/IconArrowLeftLong.vue";
import IconClock from "@/components/icons/IconClock.vue";
import IconDotsVertical from "@/components/icons/IconDotsVertical.vue";
import IconGrid from "@/components/icons/IconGrid.vue";
import IconPlus from "@/components/icons/IconPlus.vue";
import IconTrash from "@/components/icons/IconTrash.vue";
import { useApiService } from "@/composables/useApiService";
import { useOrganizations } from "~/composables/useOrganizations";

const router = useRouter();
const apiService = useApiService();
const { state: orgState, load: loadOrgs } = useOrganizations();

const {
  data: knowledgeBasesData,
  execute: fetchKnowledgeBases,
  isLoading,
} = apiService.listSharedKnowledgeBases();

const { execute: removeKnowledgeBase } = apiService.deleteSharedKnowledgeBase();

onMounted(async () => {
  await loadOrgs();
  await fetchKnowledgeBases();
});

watch(
  () => orgState.value.currentOrgId,
  async () => {
    await fetchKnowledgeBases();
  },
);

const knowledgeBases = computed(
  () => knowledgeBasesData.value?.knowledge_bases ?? [],
);

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

const showDeleteDialog = ref(false);
const knowledgeBaseToDelete = ref<string | null>(null);
const isDeleting = ref(false);

const goToCreate = () => {
  router.push("/knowledge-bases/create");
};

const goToDetails = (id: string) => {
  router.push(`/knowledge-bases/${id}`);
};

const confirmDelete = (id: string) => {
  knowledgeBaseToDelete.value = id;
  showDeleteDialog.value = true;
};

const deleteKnowledgeBase = async () => {
  if (!knowledgeBaseToDelete.value) return;
  try {
    isDeleting.value = true;
    await removeKnowledgeBase(knowledgeBaseToDelete.value);
    await fetchKnowledgeBases();
    showDeleteDialog.value = false;
    knowledgeBaseToDelete.value = null;
  } finally {
    isDeleting.value = false;
  }
};

const cancelDelete = () => {
  showDeleteDialog.value = false;
  knowledgeBaseToDelete.value = null;
};
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">Knowledge Bases</h1>
        <p class="text-sm text-muted-foreground">
          Manage reusable knowledge sources that multiple chatbots can access.
        </p>
      </div>
      <Button class="transition-all hover:shadow-md" @click="goToCreate">
        <IconPlus class="mr-2 h-4 w-4" />
        New Knowledge Base
      </Button>
    </div>

    <div v-if="isLoading" class="flex justify-center py-10">
      <div
        class="h-12 w-12 animate-spin rounded-full border-b-2 border-primary"
      ></div>
    </div>

    <div
      v-else-if="!knowledgeBases.length"
      class="flex flex-col items-center justify-center rounded-xl border border-dashed bg-card p-10 text-center"
    >
      <div
        class="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-500 text-white shadow-sm"
      >
        <IconGrid class="h-6 w-6" />
      </div>
      <h3 class="mb-1 text-lg font-medium">No knowledge bases yet</h3>
      <p class="mb-5 max-w-sm text-muted-foreground">
        Create a shared knowledge base to reuse content across multiple
        chatbots.
      </p>
      <Button class="transition-all hover:shadow-sm" @click="goToCreate">
        <IconPlus class="mr-2 h-4 w-4" />
        Create Knowledge Base
      </Button>
    </div>

    <div v-else class="grid gap-6 md:grid-cols-2">
      <AppResourceCard
        v-for="kb in knowledgeBases"
        :key="kb.id"
        :title="kb.name"
        :description="kb.description || 'No description provided'"
        :to="`/knowledge-bases/${kb.id}`"
        link-aria-label="View knowledge base"
        icon-variant="green"
      >
        <template #icon>
          <IconGrid class="h-5 w-5" />
        </template>
        <template #meta>
          <span class="text-xs text-muted-foreground">
            Created {{ formatDate(kb.created_at) }}
          </span>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button
                variant="ghost"
                size="icon"
                class="relative z-10 ml-auto h-8 w-8"
                @click.stop
                @pointerdown.stop
              >
                <IconDotsVertical class="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              align="end"
              class="w-40"
              @click.stop
              @pointerdown.stop
            >
              <DropdownMenuItem
                variant="destructive"
                @select="() => confirmDelete(kb.id)"
              >
                <IconTrash class="h-4 w-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </template>
        <template #footer>
          <div class="flex w-full flex-wrap items-center justify-between gap-2">
            <div class="flex items-center gap-2 text-muted-foreground">
              <IconClock class="h-4 w-4" />
              <span>Last updated: {{ formatDate(kb.updated_at) }}</span>
            </div>
          </div>
        </template>
      </AppResourceCard>
    </div>

    <Dialog v-model:open="showDeleteDialog">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Delete Knowledge Base</DialogTitle>
          <DialogDescription>
            This will remove the knowledge base for all chatbots. Are you sure
            you want to continue?
          </DialogDescription>
        </DialogHeader>
        <DialogFooter class="gap-2">
          <Button
            variant="outline"
            @click="cancelDelete"
            :disabled="isDeleting"
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            @click="deleteKnowledgeBase"
            :disabled="isDeleting"
          >
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
