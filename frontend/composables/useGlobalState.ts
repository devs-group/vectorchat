import { createGlobalState } from "@vueuse/core";
import { shallowRef } from "vue";

export const useGlobalState = createGlobalState(() => {
  const hasKnowledgeBaseData = shallowRef(false);
  return { hasKnowledgeBaseData };
});
