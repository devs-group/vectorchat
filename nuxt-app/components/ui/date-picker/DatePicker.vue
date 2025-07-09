<template>
  <Popover v-model:open="isOpen">
    <PopoverTrigger as-child>
      <Button
        variant="outline"
        :class="[
          'w-full justify-start text-left font-normal',
          !modelValue && 'text-muted-foreground'
        ]"
      >
        <CalendarIcon class="mr-2 h-4 w-4" />
        {{ modelValue ? formatDate(modelValue) : placeholder }}
      </Button>
    </PopoverTrigger>
    <PopoverContent class="w-auto p-0" align="start">
      <Calendar
        v-model="selectedDate"
        :min-value="minDate"
        :max-value="maxDate"
        :disabled="disabled"
        initial-focus
      />
    </PopoverContent>
  </Popover>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { CalendarDate, type DateValue } from '@internationalized/date'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { CalendarIcon } from 'lucide-vue-next'

interface Props {
  modelValue?: DateValue
  minDate?: DateValue
  maxDate?: DateValue
  disabled?: boolean
  placeholder?: string
}

interface Emits {
  (e: 'update:modelValue', value: DateValue | undefined): void
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: undefined,
  minDate: undefined,
  maxDate: undefined,
  disabled: false,
  placeholder: 'Pick a date',
})

const emit = defineEmits<Emits>()

const isOpen = ref(false)

const selectedDate = computed({
  get: () => props.modelValue,
  set: (value) => {
    emit('update:modelValue', value)
    isOpen.value = false
  },
})

const formatDate = (date: DateValue) => {
  if (!date) return ''

  // Convert DateValue to JavaScript Date for formatting
  const jsDate = new Date(date.year, date.month - 1, date.day)
  return jsDate.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

// Helper function to convert Date to CalendarDate
const toCalendarDate = (date: Date): CalendarDate => {
  return new CalendarDate(date.getFullYear(), date.getMonth() + 1, date.getDate())
}

// Helper function to convert CalendarDate to Date
const fromCalendarDate = (calendarDate: DateValue): Date => {
  return new Date(calendarDate.year, calendarDate.month - 1, calendarDate.day)
}
</script>
