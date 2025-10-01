<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">API Settings</h1>
        <p class="text-sm text-muted-foreground">
          Manage your API keys for accessing the VectorChat API. Keep your keys
          secure and never share them publicly.
        </p>
      </div>
      <Button @click="showCreateDialog = true">
        <IconPlus class="mr-2 h-4 w-4" />
        Create API Key
      </Button>
    </div>

    <div class="rounded-lg border">
      <div class="relative w-full overflow-auto">
        <table class="w-full caption-bottom text-sm">
          <thead class="[&_tr]:border-b">
            <tr
              class="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted"
            >
              <th
                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
              >
                Client ID
              </th>
              <th
                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
              >
                Name
              </th>
              <th
                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
              >
                Created
              </th>
              <th
                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
              >
                Expires
              </th>
              <th
                class="h-12 px-4 text-right align-middle font-medium text-muted-foreground"
              >
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="[&_tr:last-child]:border-0">
            <tr
              v-for="key in apiKeys?.api_keys"
              :key="key.id"
              class="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted"
            >
              <td class="p-4 align-middle">
                <div class="flex items-center gap-2">
                  <IconKey class="h-4 w-4 text-muted-foreground" />
                  <code
                    class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm"
                  >
                    {{ key.client_id }}
                  </code>
                </div>
              </td>
              <td class="p-4 align-middle">{{ key.name ?? "—" }}</td>
              <td class="p-4 align-middle">{{ formatDate(key.created_at) }}</td>
              <td class="p-4 align-middle">
                {{ key.expires_at ? formatDate(key.expires_at) : "Never" }}
              </td>
              <td class="p-4 align-middle text-right">
                <Dialog>
                  <DialogTrigger as-child>
                    <Button
                      variant="ghost"
                      size="sm"
                      class="text-destructive hover:text-destructive"
                    >
                      Revoke
                    </Button>
                  </DialogTrigger>
                  <DialogContent class="sm:max-w-md">
                    <DialogHeader>
                      <DialogTitle>Revoke OAuth Client</DialogTitle>
                      <DialogDescription>
                        Are you sure you want to revoke this client credentials
                        pair? This action cannot be undone.
                      </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                      <DialogClose as-child>
                        <Button variant="outline"> Cancel </Button>
                      </DialogClose>
                      <Button
                        variant="destructive"
                        @click="confirmRevokeKey(key)"
                        :loading="isRevokingApiKey && revokingKeyId === key.id"
                      >
                        Revoke Client
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              </td>
            </tr>
            <tr v-if="!apiKeys || apiKeys.api_keys?.length === 0">
              <td colspan="5" class="p-8 text-center text-muted-foreground">
                <div v-if="isFetchingAPIKeys" class="flex justify-center">
                  <IconSpinner
                    class="animate-spin h-5 w-5 text-muted-foreground"
                  />
                </div>
                <div v-else-if="fetchAPIKeysError">
                  Failed to load API keys. Please try again.
                </div>
                <div v-else>
                  No API keys found. Generate your first key to get started.
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination Controls -->
      <div
        v-if="apiKeys && apiKeys.pagination.total > 0"
        class="grid grid-cols-3 items-center px-6 py-4 border-t"
      >
        <div class="flex items-center gap-2">
          <span class="text-sm text-muted-foreground">Show</span>
          <select
            v-model="pageSize"
            @change="onPageSizeChange"
            class="text-sm border rounded px-2 py-1"
          >
            <option value="5">5</option>
            <option value="10">10</option>
            <option value="25">25</option>
            <option value="50">50</option>
          </select>
          <span class="text-sm text-muted-foreground">per page</span>
        </div>
        <div class="flex justify-center">
          <Pagination
            :pagination="apiKeys.pagination"
            @page-change="onPageChange"
          />
        </div>
        <div></div>
      </div>
    </div>

    <!-- Create API Key Dialog -->
    <CreateApiKeyDialog
      v-model:open="showCreateDialog"
      :is-loading="isGeneratingAPIKey"
      @generate="handleGenerateKey"
    />

    <!-- API Key Generated Dialog -->
    <ApiKeyGeneratedDialog
      v-model:open="showGeneratedDialog"
      :credentials="generatedCredentials"
      @close="handleCloseGeneratedDialog"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogClose,
} from "@/components/ui/dialog";
import { Pagination } from "@/components/ui/pagination";
import IconKey from "@/components/icons/IconKey.vue";
import IconPlus from "@/components/icons/IconPlus.vue";
import IconSpinner from "@/components/icons/IconSpinner.vue";
import type {
  APIKey,
  APIKeyCreateResponse,
  APIKeysResponse,
} from "~/types/api";
import CreateApiKeyDialog from "./components/CreateApiKeyDialog.vue";
import ApiKeyGeneratedDialog from "./components/ApiKeyGeneratedDialog.vue";

definePageMeta({
  layout: "authenticated",
});

const apiService = useApiService();
const showCreateDialog = ref(false);
const showGeneratedDialog = ref(false);
const generatedCredentials = ref<{
  clientId: string;
  clientSecret: string;
  name?: string | null;
  expiresAt?: string | null;
} | null>(null);
const revokingKeyId = ref<string | null>(null);

// Pagination state
const currentPage = ref(1);
const pageSize = ref(10);

// HTTP calls
const {
  execute: fetchAPIKeys,
  data: apiKeys,
  isLoading: isFetchingAPIKeys,
  error: fetchAPIKeysError,
} = apiService.listApiKeys<APIKeysResponse>();
const {
  execute: generateApiKey,
  data: newKey,
  isLoading: isGeneratingAPIKey,
  error: generateApiKeyError,
} = apiService.generateApiKey<APIKeyCreateResponse>();
const {
  execute: revokeApiKey,
  isLoading: isRevokingApiKey,
  error: revokeApiKeyError,
} = apiService.revokeApiKey();

// Format date for display
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

// Pagination handlers
const onPageChange = async (page: number) => {
  currentPage.value = page;
  await loadAPIKeys();
};

const onPageSizeChange = async () => {
  currentPage.value = 1; // Reset to first page when changing page size
  await loadAPIKeys();
};

// Load API keys with current pagination
const loadAPIKeys = async () => {
  await fetchAPIKeys(currentPage.value, pageSize.value);
};

// Handle generate key from dialog
const handleGenerateKey = async (data: {
  name: string;
  expires_at?: string;
}) => {
  try {
    await generateApiKey(data);

    if (newKey.value?.client_secret && newKey.value?.client_id) {
      generatedCredentials.value = {
        clientId: newKey.value.client_id,
        clientSecret: newKey.value.client_secret,
        name: newKey.value.name ?? null,
        expiresAt: newKey.value.expires_at ?? null,
      };
      showCreateDialog.value = false;
      showGeneratedDialog.value = true;

      // Refresh the list after generating a new key
      await loadAPIKeys();
    } else {
      console.error("API key not found in response structure");
    }
  } catch (err) {
    console.error("Failed to generate API key:", err);
  }
};

// Handle close of generated dialog
const handleCloseGeneratedDialog = () => {
  generatedCredentials.value = null;
  showGeneratedDialog.value = false;
};

// Confirm revoke key
const confirmRevokeKey = async (key: APIKey) => {
  try {
    revokingKeyId.value = key.id;
    await revokeApiKey(key.id);
    await loadAPIKeys();
  } finally {
    revokingKeyId.value = null;
  }
};

// Fetch API keys on mount
onMounted(async () => {
  await loadAPIKeys();
});
</script>
