<script setup lang="ts">
definePageMeta({
  layout: "authenticated",
});
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'

const router = useRouter()
const apiService = useApiService()

const name = ref('')
const description = ref('')
const systemInstructions = ref('You are a helpful AI assistant')
const modelName = ref('gpt-4')
const temperatureParam = ref(0.7)
const maxTokens = ref(2000)

const isLoading = ref(false)

const handleSubmit = async () => {
  if (!name.value.trim()) {
    toast.error('Name is required')
    return
  }
  isLoading.value = true
  try {
    const { execute, data, error } = apiService.createChatbot({
      name: name.value,
      description: description.value,
      model_name: modelName.value,
      system_instructions: systemInstructions.value,
      max_tokens: Number(maxTokens.value),
      temperature_param: Number(temperatureParam.value),
    })
    await execute()
    if (error.value) throw error.value
    toast.success('Chatbot created successfully!')
    router.push('/chat')
  } catch (err: any) {
    toast.error('Failed to create chatbot', {
      description: err?.message || 'An error occurred',
    })
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen bg-background px-4 py-12">
    <div class="flex flex-col w-full max-w-xl justify-start">
      <h1 class="mb-8 text-2xl font-bold tracking-tight text-left">Create a New Chatbot</h1>
      <form @submit.prevent="handleSubmit" class="space-y-6 w-full">
        <div>
          <Label for="name">Name <span class="text-destructive">*</span></Label>
          <Input id="name" v-model="name" placeholder="My AI Assistant" required class="mt-2" />
        </div>
        <div>
          <Label for="description">Description</Label>
          <Textarea id="description" v-model="description" placeholder="A helpful AI assistant for my project" class="mt-2 min-h-[80px]" />
        </div>
        <div>
          <Label for="systemInstructions">System Instructions</Label>
          <Textarea id="systemInstructions" v-model="systemInstructions" placeholder="You are a helpful AI assistant" class="mt-2 min-h-[80px]" />
        </div>
        <div>
          <Label for="modelName">Model Name</Label>
          <Input id="modelName" v-model="modelName" placeholder="gpt-4" class="mt-2" />
        </div>
        <div class="flex gap-4">
          <div class="flex-1">
            <Label for="temperatureParam">Temperature</Label>
            <Input id="temperatureParam" v-model="temperatureParam" type="number" step="0.01" min="0" max="2" placeholder="0.7" class="mt-2" />
          </div>
          <div class="flex-1">
            <Label for="maxTokens">Max Tokens</Label>
            <Input id="maxTokens" v-model="maxTokens" type="number" min="1" max="4000" placeholder="2000" class="mt-2" />
          </div>
        </div>
        <Button type="submit" :loading="isLoading" class="mt-4 w-full sm:w-auto">Create Chatbot</Button>
      </form>
    </div>
    <!-- Empty right side for future content -->
    <div class="flex-1"></div>
  </div>
</template>
