import type { SessionResponse } from "~/types/api";

export function useSession() {
  const state = useState<SessionResponse | null>("session", () => null);

  const call = async () => {
    const api = useApiService().getSession();
    await api.execute();

    if (api.error.value || !api.data.value) {
      state.value = null;
      return { ok: false as const };
    }

    state.value = api.data.value as SessionResponse;
    return { ok: true as const, data: state.value };
  };

  const get = () => state;

  return { call, get };
}
