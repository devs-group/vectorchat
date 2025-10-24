import { defineEventHandler } from "h3";

export default defineEventHandler(async (event) => {
  const availableWidgets = [
    {
      name: "vectorchat-plex-widget",
      description: "Monospaced, dashboard-inspired style with IBM Plex Mono vibe",
      theme: "dark"
    },
    {
      name: "vectorchat-glass-widget",
      description: "Glass morphism design with frosted glass effects",
      theme: "glass"
    }
  ];

  return {
    widgets: availableWidgets,
    message: "Available VectorChat widgets"
  };
});
