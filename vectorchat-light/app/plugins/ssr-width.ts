import { provideSSRWidth } from "@vueuse/core";

export default defineNuxtPlugin(() => {
  provideSSRWidth();
});
