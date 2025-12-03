<template>
  <div class="grid gap-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold">Organizations</h1>
        <p class="text-sm text-muted-foreground">
          Manage organizations.
        </p>
      </div>
      <Button @click="openCreate = true">
        <span class="hidden sm:inline">Create organization</span>
        <span class="sm:hidden">New</span>
      </Button>
    </div>

    <Card>
      <CardHeader>
        <CardTitle class="text-base">Your workspaces</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="divide-y divide-border">
          <div
            v-for="org in orgs"
            :key="org.id"
            class="flex flex-col gap-1 py-3 sm:flex-row sm:items-center sm:justify-between"
          >
            <div class="flex items-center gap-3">
              <div
                class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-sm font-semibold uppercase text-primary"
              >
                {{ initials(org.name) }}
              </div>
              <div class="flex flex-col">
                <span class="font-medium">{{ org.name }}</span>
                <span class="text-xs text-muted-foreground">
                  {{ org.role }}
                </span>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <Badge v-if="isCurrent(org)" variant="secondary">Active</Badge>
              <Button size="sm" variant="ghost" @click="selectOrg(org)">
                {{ isCurrent(org) ? "Using" : "Use" }}
              </Button>
              <Button
                v-if="canInvite(org)"
                size="sm"
                variant="outline"
                @click="openInviteDialog(org)"
              >
                Invite
              </Button>
            </div>
          </div>
          <div
            v-if="orgs.length === 0"
            class="py-8 text-center text-sm text-muted-foreground"
          >
            No organizations yet. Create one to collaborate.
          </div>
        </div>
      </CardContent>
    </Card>

    <Dialog v-model:open="openCreate">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create organization</DialogTitle>
          <DialogDescription>
            Give your team a name. You can change it later.
          </DialogDescription>
        </DialogHeader>
        <form class="space-y-4" @submit.prevent="submit">
          <div class="space-y-2">
            <Label for="org-name">Name</Label>
            <Input
              id="org-name"
              v-model="form.name"
              placeholder="Acme Inc"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="org-desc">Description</Label>
            <Input
              id="org-desc"
              v-model="form.description"
              placeholder="Team workspace"
            />
          </div>
          <div class="flex justify-end gap-2">
            <Button type="button" variant="ghost" @click="openCreate = false">
              Cancel
            </Button>
            <Button type="submit" :disabled="creating">Create</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="openInvite">
      <DialogContent class="max-w-lg">
        <DialogHeader>
          <DialogTitle>
            Invite to {{ inviteOrg?.name || "organization" }}
          </DialogTitle>
          <DialogDescription>
            Send an invitation to join this workspace.
          </DialogDescription>
        </DialogHeader>
        <form class="space-y-4" @submit.prevent="submitInvite">
          <div class="space-y-2">
            <Label for="invite-email">Email</Label>
            <Input
              id="invite-email"
              v-model="inviteForm.email"
              type="email"
              placeholder="person@example.com"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="invite-role">Role</Label>
            <Select v-model="inviteForm.role">
              <SelectTrigger id="invite-role" class="w-full">
                <SelectValue placeholder="Choose a role" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="member">Member</SelectItem>
                <SelectItem value="admin">Admin</SelectItem>
                <SelectItem value="billing">Billing</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="space-y-2">
            <Label for="invite-message">Message</Label>
            <Textarea
              id="invite-message"
              v-model="inviteForm.message"
              rows="3"
              placeholder="Optional note for the recipient"
            />
          </div>
          <div class="flex justify-end gap-2">
            <Button
              type="button"
              variant="ghost"
              @click="openInvite = false"
            >
              Cancel
            </Button>
            <Button type="submit" :disabled="inviting">
              {{ inviting ? "Sending..." : "Send invite" }}
            </Button>
          </div>
        </form>

        <div
          v-if="inviteToken"
          class="rounded-md border border-dashed border-primary/40 bg-primary/5 p-3"
        >
          <p class="text-xs text-muted-foreground">
            Share this token with the invitee to accept:
          </p>
          <div class="mt-2 flex items-center gap-2">
            <Input :value="inviteToken" readonly class="font-mono text-xs" />
            <Button type="button" variant="ghost" size="sm" @click="copyToken">
              Copy
            </Button>
          </div>
        </div>

        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <p class="text-sm font-medium">Pending invites</p>
            <span class="text-xs text-muted-foreground" v-if="inviteOrg">
              {{ inviteOrg.name }}
            </span>
          </div>
          <div v-if="loadingInvites" class="text-sm text-muted-foreground">
            Loading invites...
          </div>
          <div
            v-else-if="invites.length === 0"
            class="text-sm text-muted-foreground"
          >
            No invites yet.
          </div>
          <div v-else class="divide-y divide-border">
            <div
              v-for="invite in invites"
              :key="invite.id"
              class="flex items-center justify-between py-2 text-sm"
            >
              <div class="space-y-0.5">
                <div class="font-medium">{{ invite.email }}</div>
                <div class="text-xs text-muted-foreground">
                  {{ invite.role }} · expires {{ expiresLabel(invite.expires_at) }}
                </div>
              </div>
              <Badge
                v-if="invite.accepted_at"
                variant="secondary"
                class="text-xs"
              >
                Accepted
              </Badge>
              <Badge v-else variant="outline" class="text-xs">
                Pending
              </Badge>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { toast } from "vue-sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { useOrganizations } from "~/composables/useOrganizations";
import { useApiService } from "@/composables/useApiService";
import type { Organization, OrganizationInvite } from "~/types/api";

definePageMeta({
  layout: "authenticated",
});

const { state, load, setCurrent, create } = useOrganizations();
const openCreate = ref(false);
const creating = ref(false);
const form = reactive({
  name: "",
  description: "",
});
const openInvite = ref(false);
const inviteOrg = ref<Organization | null>(null);
const inviteToken = ref<string | null>(null);
const inviteForm = reactive({
  email: "",
  role: "member",
  message: "",
});

const api = useApiService();
const {
  execute: sendInvite,
  data: inviteResponse,
  isLoading: inviting,
} = api.createOrganizationInvite();
const {
  execute: fetchInvites,
  data: invitesData,
  isLoading: loadingInvites,
} = api.listOrganizationInvites();

const orgs = computed(() => state.value.organizations);
const invites = computed<OrganizationInvite[]>(() => {
  const res = invitesData.value as { invites?: OrganizationInvite[] } | null;
  return res?.invites ?? [];
});

const initials = (name: string) =>
  name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? "")
    .join("") || "VC";

const submit = async () => {
  if (!form.name.trim()) return;
  creating.value = true;
  try {
    await create({
      name: form.name.trim(),
      description: form.description || undefined,
    });
    toast.success("Organization created");
    openCreate.value = false;
    form.name = "";
    form.description = "";
  } catch (err: any) {
    toast.error("Failed to create organization", {
      description: err?.message ?? "Unexpected error",
    });
  } finally {
    creating.value = false;
    await load();
  }
};

const isCurrent = (org: Organization) =>
  state.value.currentOrgId === null
    ? org.id === "00000000-0000-0000-0000-000000000000"
    : org.id === state.value.currentOrgId;

const selectOrg = (org: Organization) => {
  setCurrent(org);
  toast.success(`Switched to ${org.name}`);
};

const canInvite = (org: Organization) =>
  org.id !== "00000000-0000-0000-0000-000000000000" &&
  ["owner", "admin"].includes(org.role);

const openInviteDialog = async (org: Organization) => {
  inviteOrg.value = org;
  inviteToken.value = null;
  inviteForm.email = "";
  inviteForm.message = "";
  inviteForm.role = "member";
  openInvite.value = true;
  await fetchInvites(org.id);
};

const submitInvite = async () => {
  if (!inviteOrg.value || !inviteForm.email.trim()) return;
  await sendInvite({
    organizationId: inviteOrg.value.id,
    payload: {
      email: inviteForm.email.trim(),
      role: inviteForm.role,
      message: inviteForm.message || undefined,
    },
  });
  const payload = inviteResponse.value as
    | { token?: string; invite?: OrganizationInvite }
    | null;
  if (payload?.token) {
    inviteToken.value = payload.token;
  }
  if (inviteOrg.value) {
    await fetchInvites(inviteOrg.value.id);
  }
};

const copyToken = async () => {
  if (!inviteToken.value) return;
  try {
    await navigator.clipboard.writeText(inviteToken.value);
    toast.success("Token copied");
  } catch (err: any) {
    toast.error("Failed to copy token", {
      description: err?.message ?? "Unexpected error",
    });
  }
};

const expiresLabel = (date: string) =>
  new Date(date).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
  });

onMounted(async () => {
  await load();
});
</script>
