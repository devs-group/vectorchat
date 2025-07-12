<template>
  <div class="bg-card border border-border rounded-lg p-6 mt-12">
    <div class="mb-6">
      <h3 class="text-xl font-semibold mb-2">Chat API Documentation</h3>
      <p class="text-sm text-muted-foreground">
        Use your API key to integrate this chatbot into your applications
      </p>
    </div>

    <!-- API Endpoint Info -->
    <div class="mb-6 p-4 bg-muted rounded-lg">
      <div class="flex items-center mb-2">
        <span
          class="bg-green-100 text-green-800 text-xs font-medium px-2.5 py-0.5 rounded mr-3"
        >
          POST
        </span>
        <code class="text-sm font-mono bg-background px-2 py-1 rounded">
          {{ baseUrl }}/api/chat/{{ chatId }}/message
        </code>
      </div>
      <p class="text-sm text-muted-foreground">
        Send a message to your chatbot and receive an AI-powered response
      </p>
    </div>

    <!-- cURL Example -->
    <div class="mb-6">
      <h4 class="font-medium mb-3">Example Request</h4>
      <div class="bg-slate-900 text-slate-100 p-4 rounded-lg overflow-x-auto">
        <pre
          class="text-sm"
        ><code>curl -X POST {{ baseUrl }}/api/chat/{{ chatId }}/message \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "query": "What is this project about?"
  }'</code></pre>
      </div>
    </div>

    <!-- Request/Response Documentation -->
    <div class="space-y-6">
      <div>
        <h4 class="font-medium mb-3">Request Body</h4>
        <div class="bg-muted p-4 rounded-lg">
          <code class="text-sm">
            {<br />
            &nbsp;&nbsp;"query": "string"<br />
            }
          </code>
        </div>
      </div>

      <div>
        <h4 class="font-medium mb-3">Response</h4>
        <div class="bg-muted p-4 rounded-lg">
          <code class="text-sm">
            {<br />
            &nbsp;&nbsp;"response": "string"<br />
            }
          </code>
        </div>
      </div>

      <div>
        <h4 class="font-medium mb-3">Authentication</h4>
        <div class="bg-amber-50 border border-amber-200 p-4 rounded-lg">
          <p class="text-sm text-amber-800 mb-2">
            Include your API key in the X-API-Key header.
          </p>
          <code class="text-sm bg-amber-100 px-2 py-1 rounded text-amber-900">
            X-API-Key: YOUR_API_KEY
          </code>
          <p class="text-sm text-amber-700 mt-2">
            <NuxtLink
              to="/settings"
              class="text-amber-800 hover:underline font-medium"
            >
              Create an API key in Settings →
            </NuxtLink>
          </p>
        </div>
      </div>

      <div>
        <h4 class="font-medium mb-3">Error Responses</h4>
        <div class="space-y-3">
          <div class="bg-muted p-4 rounded-lg">
            <div class="flex items-center mb-2">
              <span
                class="bg-red-100 text-red-800 text-xs font-medium px-2 py-1 rounded mr-2"
              >
                400
              </span>
              <span class="text-sm font-medium">Bad Request</span>
            </div>
            <code class="text-sm text-muted-foreground">
              Missing or invalid query parameter
            </code>
          </div>

          <div class="bg-muted p-4 rounded-lg">
            <div class="flex items-center mb-2">
              <span
                class="bg-red-100 text-red-800 text-xs font-medium px-2 py-1 rounded mr-2"
              >
                401
              </span>
              <span class="text-sm font-medium">Unauthorized</span>
            </div>
            <code class="text-sm text-muted-foreground">
              Invalid or missing API key
            </code>
          </div>

          <div class="bg-muted p-4 rounded-lg">
            <div class="flex items-center mb-2">
              <span
                class="bg-red-100 text-red-800 text-xs font-medium px-2 py-1 rounded mr-2"
              >
                404
              </span>
              <span class="text-sm font-medium">Not Found</span>
            </div>
            <code class="text-sm text-muted-foreground">
              Chatbot not found
            </code>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  chatId: string;
}

const props = defineProps<Props>();

// Get base URL from runtime config or use current origin
const config = useRuntimeConfig();
const baseUrl = computed(() => {
  if (process.client) {
    return window.location.origin;
  }
  return config.public.baseUrl || "https://your-domain.com";
});
</script>

<style scoped>
pre {
  white-space: pre-wrap;
  word-wrap: break-word;
}

code {
  font-family: "Monaco", "Menlo", "Ubuntu Mono", monospace;
}
</style>
