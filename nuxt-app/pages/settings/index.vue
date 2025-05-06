<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <h1 class="text-3xl font-bold tracking-tight">API Settings</h1>
      <Dialog>
        <DialogTrigger as-child>
          <Button>
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="mr-2 h-4 w-4"
            >
              <path d="M5 12h14"></path>
              <path d="M12 5v14"></path>
            </svg>
            Create API Key
          </Button>
        </DialogTrigger>
        <DialogContent class="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Generate New API Key</DialogTitle>
            <DialogDescription>
              Generate a new API key to access the VectorChat API. Keep your keys
              secure and never share them publicly.
            </DialogDescription>
          </DialogHeader>
          <div v-show="!newKey?.api_key.key" class="flex flex-col gap-4 py-4">
            <div class="flex items-center gap-2">
              <Input
                id="key"
                v-model="newKeyName"
                placeholder="Enter a name for your API key"
              />
            </div>
          </div>
          <div v-show="newKey?.api_key.key" class="flex items-center space-x-2">
          <div class="grid flex-1 gap-2">
            <Label for="key">API Key</Label>
            <Input id="key" :value="newKey?.api_key.key" readonly class="font-mono" />
          </div>
          <Button
            type="submit"
            size="sm"
            class="px-3"
            @click="copyToClipboard(newKey?.key)"
          >
            <span class="sr-only">Copy</span>
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="h-4 w-4"
            >
              <rect width="14" height="14" x="8" y="8" rx="2" ry="2"></rect>
              <path
                d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"
              ></path>
            </svg>
          </Button>
        </div>
          <DialogFooter>
            <DialogClose as-child>
              <Button variant="outline">
                Cancel
              </Button>
            </DialogClose>
            <Button @click="generateNewKey" :loading="isGeneratingAPIKey">
              Generate Key
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>

    <div class="rounded-lg border">
      <div class="p-6">
        <h2 class="text-lg font-semibold">Your API Keys</h2>
        <p class="text-sm text-muted-foreground">
          Manage your API keys for accessing the VectorChat API. Keep your keys
          secure and never share them publicly.
        </p>
      </div>
      <div class="relative w-full overflow-auto">
        <table class="w-full caption-bottom text-sm">
          <thead class="[&_tr]:border-b">
            <tr
              class="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted"
            >
              <th
                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
              >
                Name
              </th>
              <th
                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
              >
                Key
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
                class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
              >
                Status
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
              <td class="p-4 align-middle">{{ key.name }}</td>
              <td class="p-4 align-middle">
                <div class="flex items-center gap-2">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="24"
                    height="24"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    class="h-4 w-4 text-muted-foreground"
                  >
                    <path d="M15 7h3a2 2 0 0 1 2 2v6a2 2 0 0 1-2 2h-3"></path>
                    <path d="M10 17H7a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2h3"></path>
                    <line x1="8" x2="16" y1="12" y2="12"></line>
                  </svg>
                  <code
                    class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm"
                  >
                    {{ key.key }}
                  </code>
                </div>
              </td>
              <td class="p-4 align-middle">{{ formatDate(key.created_at) }}</td>
              <td class="p-4 align-middle">{{ formatDate(key.expires_at) }}</td>
              <td class="p-4 align-middle">
                <span
                  :class="[
                    'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold',
                    key.revoked_at
                      ? 'bg-destructive/10 text-destructive'
                      : 'bg-green-100 text-green-800',
                  ]"
                >
                  {{ key.revoked_at ? "Revoked" : "Active" }}
                </span>
              </td>
              <td class="p-4 align-middle text-right">
                <Dialog v-if="!key.revoked_at">
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
                      <DialogTitle>Revoke API Key</DialogTitle>
                      <DialogDescription>
                        Are you sure you want to revoke this API key? This action cannot be
                        undone.
                      </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                      <DialogClose as-child>
                        <Button variant="outline">
                          Cancel
                        </Button>
                      </DialogClose>
                      <Button
                        variant="destructive"
                        @click="confirmRevokeKey(key)"
                        :loading="isRevokingApiKey && revokingKeyId === key.id"
                      >
                        Revoke Key
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              </td>
            </tr>
            <tr v-if="!apiKeys || apiKeys.length === 0">
              <td colspan="6" class="p-8 text-center text-muted-foreground">
                <div v-if="isFetchingAPIKeys" class="flex justify-center">
                  <svg class="animate-spin h-5 w-5 text-muted-foreground" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogClose,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { APIKeyResponse, APIKeysResponse } from '~/types/api'

definePageMeta({
  layout: "authenticated",
});

interface APIKey {
  id: string;
  key: string;
  name: string;
  created_at: string;
  expires_at: string;
  revoked_at: string | null;
  user_id: string;
}

const apiService = useApiService()
const newKeyName = ref("");
const showNewKeyDialog = ref(false);
const revokingKeyId = ref<string | null>(null);

// HTTP calls
const { execute: fetchAPIKeys, data: apiKeys, isLoading: isFetchingAPIKeys, error: fetchAPIKeysError } = apiService.listApiKeys<APIKeysResponse>()
const { execute: generateApiKey, data: newKey, isLoading: isGeneratingAPIKey, error: generateApiKeyError } = apiService.generateApiKey<APIKeyResponse>()
const { execute: revokeApiKey, isLoading: isRevokingApiKey, error: revokeApiKeyError } = apiService.revokeApiKey()

// Format date for display
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

// Copy to clipboard
const copyToClipboard = async (text: string | undefined) => {
  if (!text) return;
  try {
    await navigator.clipboard.writeText(text);
    // TODO: Show toast notification
  } catch (err) {
    console.error("Failed to copy text: ", err);
  }
};

// Generate new API key
const generateNewKey = async () => {
  if (!newKeyName.value) return;
  
  try {
    await generateApiKey({ name: newKeyName.value });
    showNewKeyDialog.value = true;
    newKeyName.value = "";
    
    // Refresh the list after generating a new key
    await fetchAPIKeys();
  } catch (err) {
    console.error("Failed to generate API key:", err);
  }
};

// Confirm revoke key
const confirmRevokeKey = async (key: APIKey) => {
  await revokeApiKey(key.id);
  await fetchAPIKeys();
};

// Fetch API keys on mount
onMounted(async () => {
  await fetchAPIKeys();
});
</script>