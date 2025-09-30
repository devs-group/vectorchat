import { watch } from "vue";

import { useKratosSession } from "~/composables/useKratosSession";
import type { KratosSession } from "~/composables/useKratosSession";
import type { SessionResponse } from "~/types/api";

const toSessionResponse = (session: KratosSession): SessionResponse => {
  const identity = session.identity;
  const traits = identity.traits ?? {};
  const now = new Date().toISOString();

  return {
    user: {
      id: identity.id,
      name: (traits.name as string) || (traits.email as string) || "",
      email: (traits.email as string) || "",
      provider:
        (identity.metadata_public?.provider as string) ||
        (traits.provider as string) ||
        "github",
      created_at:
        (identity.created_at as string | undefined) ||
        (session.authenticated_at as string | undefined) ||
        now,
      updated_at:
        (identity.updated_at as string | undefined) ||
        (session.authenticated_at as string | undefined) ||
        now,
    },
  };
};

export function useSession() {
  const state = useState<SessionResponse | null>("session", () => null);
  const { session, loadSession } = useKratosSession();

  const call = async () => {
    const current = await loadSession();
    if (!current) {
      state.value = null;
      return { ok: false as const };
    }

    state.value = toSessionResponse(current);
    return { ok: true as const, data: state.value };
  };

  const get = () => state;

  watch(
    session,
    (value) => {
      state.value = value ? toSessionResponse(value) : null;
    },
    { immediate: true },
  );

  return { call, get };
}
