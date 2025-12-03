import { useStorage } from "@vueuse/core";

import type {
  Organization,
  OrganizationCreateRequest,
  OrganizationListResponse,
} from "~/types/api";

type OrgState = {
  currentOrgId: string | null;
  currentOrgName: string;
  currentRole: string;
  organizations: Organization[];
};

const STORAGE_KEY = "vectorchat-org-id";

export function useOrganizations() {
  const persistedOrgId = useStorage<string | null>(STORAGE_KEY, null);
  const state = useState<OrgState>("org-state", () => ({
    currentOrgId: null,
    currentOrgName: "Personal",
    currentRole: "personal",
    organizations: [],
  }));

  const setCurrent = (org: Organization | null) => {
    if (org && org.id !== "00000000-0000-0000-0000-000000000000") {
      state.value.currentOrgId = org.id;
      state.value.currentOrgName = org.name;
      state.value.currentRole = org.role;
      persistedOrgId.value = org.id;
    } else {
      state.value.currentOrgId = null;
      state.value.currentOrgName = "Personal";
      state.value.currentRole = "personal";
      persistedOrgId.value = null;
    }
  };

  const ensureSelection = (organizations: Organization[]) => {
    // Personal workspace stub (id nil) is always first in backend response; fall back when not found
    const stored = persistedOrgId.value;
    if (stored) {
      const match = organizations.find((o) => o.id === stored);
      if (match) {
        setCurrent(match);
        return;
      }
    }
    // default to personal (id nil) or first org
    const personal = organizations.find(
      (o) => o.id === "00000000-0000-0000-0000-000000000000",
    );
    setCurrent(personal ?? organizations[0] ?? null);
  };

  const load = async () => {
    const { listOrganizations } = useApiService();
    const { data, execute } = listOrganizations<OrganizationListResponse>();
    await execute();
    if (data.value?.organizations) {
      state.value.organizations = data.value.organizations;
      ensureSelection(data.value.organizations);
    }
  };

  const create = async (payload: OrganizationCreateRequest) => {
    const { createOrganization } = useApiService();
    const { data, execute } = createOrganization<Organization>();
    await execute(payload);
    if (data.value) {
      // refresh list afterwards
      await load();
      setCurrent(
        state.value.organizations.find((o) => o.id === data.value!.id) ??
          data.value,
      );
    }
  };

  const currentOrgHeader = computed(() =>
    state.value.currentOrgId ? { "X-Organization-ID": state.value.currentOrgId } : {},
  );

  return {
    state,
    load,
    create,
    setCurrent,
    currentOrgHeader,
  };
}
