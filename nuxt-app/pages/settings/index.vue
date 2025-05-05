<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <h1 class="text-3xl font-bold tracking-tight">API Settings</h1>
      <Button @click="showGenerateKeyDialog = true">
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
      </Button>
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
              v-for="key in apiKeys"
              :key="key.id"
              class="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted"
            >
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
                <Button
                  v-if="!key.revoked_at"
                  variant="ghost"
                  size="sm"
                  class="text-destructive hover:text-destructive"
                  @click="showRevokeDialog(key)"
                  :loading="isRevoking === key.id"
                >
                  Revoke
                </Button>
              </td>
            </tr>
            <tr v-if="apiKeys.length === 0">
              <td colspan="5" class="p-8 text-center text-muted-foreground">
                No API keys found. Generate your first key to get started.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Generate Key Dialog -->
    <Dialog v-model="showGenerateKeyDialog">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Generate New API Key</DialogTitle>
          <DialogDescription>
            Generate a new API key to access the VectorChat API. Keep your keys
            secure and never share them publicly.
          </DialogDescription>
        </DialogHeader>
        <div class="flex flex-col gap-4 py-4">
          <div class="flex items-center gap-2">
            <Input
              id="key"
              v-model="newKeyName"
              placeholder="Enter a name for your API key"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showGenerateKeyDialog = false">
            Cancel
          </Button>
          <Button @click="generateNewKey" :loading="isGenerating">
            Generate Key
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- New Key Dialog -->
    <Dialog v-model="showNewKeyDialog">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>New API Key Generated</DialogTitle>
          <DialogDescription>
            Your new API key has been generated. Make sure to copy it now. You
            won't be able to see it again!
          </DialogDescription>
        </DialogHeader>
        <div class="flex items-center space-x-2">
          <div class="grid flex-1 gap-2">
            <Label for="key">API Key</Label>
            <Input id="key" :value="newKey?.key" readonly class="font-mono" />
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
        <DialogFooter class="sm:justify-start">
          <Button
            type="button"
            variant="secondary"
            @click="showNewKeyDialog = false"
          >
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Revoke Key Dialog -->
    <Dialog v-model="showRevokeKeyDialog">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Revoke API Key</DialogTitle>
          <DialogDescription>
            Are you sure you want to revoke this API key? This action cannot be
            undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="showRevokeKeyDialog = false">
            Cancel
          </Button>
          <Button
            variant="destructive"
            @click="confirmRevokeKey"
            :loading="isRevoking === keyToRevoke?.id"
          >
            Revoke Key
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: "authenticated",
});

interface APIKey {
  id: string;
  key: string;
  created_at: string;
  expires_at: string;
  revoked_at: string | null;
  user_id: string;
}

const apiKeys = ref<APIKey[]>([]);
const isGenerating = ref(false);
const isRevoking = ref<string | null>(null);
const showGenerateKeyDialog = ref(false);
const showNewKeyDialog = ref(false);
const showRevokeKeyDialog = ref(false);
const newKey = ref<APIKey | null>(null);
const keyToRevoke = ref<APIKey | null>(null);
const newKeyName = ref("");

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

// Fetch API keys
const fetchAPIKeys = async () => {
  try {
    const response = await fetch("/api/auth/apikey");
    const data = await response.json();
    apiKeys.value = data.api_keys;
  } catch (error) {
    console.error("Error fetching API keys:", error);
    // TODO: Show error toast
  }
};

// Generate new API key
const generateNewKey = async () => {
  isGenerating.value = true;
  try {
    const response = await fetch("/api/auth/apikey", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name: newKeyName.value }),
    });
    const data = await response.json();
    newKey.value = data.api_key;
    showGenerateKeyDialog.value = false;
    showNewKeyDialog.value = true;
    newKeyName.value = "";
    await fetchAPIKeys();
  } catch (error) {
    console.error("Error generating API key:", error);
    // TODO: Show error toast
  } finally {
    isGenerating.value = false;
  }
};

// Show revoke dialog
const showRevokeDialog = (key: APIKey) => {
  keyToRevoke.value = key;
  showRevokeKeyDialog.value = true;
};

// Confirm revoke key
const confirmRevokeKey = async () => {
  if (!keyToRevoke.value) return;
  isRevoking.value = keyToRevoke.value.id;
  try {
    await fetch(`/api/auth/apikey/${keyToRevoke.value.id}`, {
      method: "DELETE",
    });
    showRevokeKeyDialog.value = false;
    keyToRevoke.value = null;
    await fetchAPIKeys();
    // TODO: Show success toast
  } catch (error) {
    console.error("Error revoking API key:", error);
    // TODO: Show error toast
  } finally {
    isRevoking.value = null;
  }
};

// Fetch API keys on mount
onMounted(() => {
  fetchAPIKeys();
});
</script>
