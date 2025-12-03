<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button
        variant="ghost"
        class="w-full justify-between px-3 text-left font-medium"
      >
        <div class="flex items-center gap-3 overflow-hidden">
          <div
            class="flex h-8 w-8 items-center justify-center rounded-md bg-primary/10 text-sm font-semibold uppercase text-primary"
          >
            {{ initials }}
          </div>
          <div class="flex flex-col leading-tight overflow-hidden">
            <span class="truncate text-sm font-semibold">
              {{ currentName }}
            </span>
            <span class="truncate text-xs text-muted-foreground">
              {{ currentRole }}
            </span>
          </div>
        </div>
        <IconChevronDown class="h-4 w-4 opacity-70" />
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent class="w-64">
      <DropdownMenuLabel class="text-xs uppercase text-muted-foreground">
        Workspaces
      </DropdownMenuLabel>
      <DropdownMenuSeparator />
      <DropdownMenuRadioGroup
        :model-value="currentOrgId || 'personal'"
        @update:model-value="onSelect"
      >
        <DropdownMenuRadioItem value="personal">
          Personal workspace
        </DropdownMenuRadioItem>
        <DropdownMenuRadioItem
          v-for="org in filteredOrgs"
          :key="org.id"
          :value="org.id"
        >
          <div class="flex flex-col">
            <span class="text-sm font-medium leading-tight">{{ org.name }}</span>
            <span class="text-xs text-muted-foreground">{{ org.role }}</span>
          </div>
        </DropdownMenuRadioItem>
      </DropdownMenuRadioGroup>
      <DropdownMenuSeparator />
    </DropdownMenuContent>
  </DropdownMenu>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import IconChevronDown from "@/components/icons/IconChevronDown.vue";
import {
  Button,
} from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { toast } from "vue-sonner";
import { useOrganizations } from "~/composables/useOrganizations";

const { state, setCurrent, load } = useOrganizations();

const currentOrgId = computed(() => state.value.currentOrgId);
const currentName = computed(() => state.value.currentOrgName);
const currentRole = computed(() => state.value.currentRole);
const filteredOrgs = computed(() =>
  state.value.organizations.filter(
    (o) => o.id !== "00000000-0000-0000-0000-000000000000",
  ),
);

const initials = computed(() => {
  const name = currentName.value;
  if (!name) return "VC";
  const parts = name.split(" ").filter(Boolean);
  return parts
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? "")
    .join("");
});

const redirectToChat = () => {
  window.location.assign("/chat");
};

const onSelect = (val: string) => {
  const personal = val === "personal";
  if (personal) {
    setCurrent(null);
    toast.success("Switched to Personal workspace");
    redirectToChat();
    return;
  }
  const target = state.value.organizations.find((o) => o.id === val);
  if (target) {
    setCurrent(target);
    toast.success(`Switched to ${target.name}`);
    redirectToChat();
  }
};

onMounted(async () => {
  await load();
});
</script>
