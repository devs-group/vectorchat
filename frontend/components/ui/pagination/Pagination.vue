<template>
  <nav
    :aria-label="$attrs['aria-label'] ?? 'pagination'"
    role="navigation"
    class="mx-auto flex w-full justify-center"
  >
    <div class="flex items-center space-x-2">
      <Button
        variant="outline"
        size="sm"
        :disabled="!pagination.has_prev"
        @click="$emit('page-change', pagination.page - 1)"
      >
        <ChevronLeft class="h-4 w-4" />
        <span class="sr-only">Previous</span>
      </Button>

      <div class="flex items-center space-x-1">
        <template v-for="page in visiblePages" :key="page">
          <Button
            v-if="page !== '...'"
            variant="outline"
            size="sm"
            :class="[
              page === pagination.page
                ? 'bg-primary text-primary-foreground hover:bg-primary/90'
                : 'hover:bg-accent hover:text-accent-foreground'
            ]"
            @click="$emit('page-change', page)"
          >
            {{ page }}
          </Button>
          <div v-else class="px-2 py-1 text-sm text-muted-foreground">
            ...
          </div>
        </template>
      </div>

      <Button
        variant="outline"
        size="sm"
        :disabled="!pagination.has_next"
        @click="$emit('page-change', pagination.page + 1)"
      >
        <span class="sr-only">Next</span>
        <ChevronRight class="h-4 w-4" />
      </Button>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import type { PaginationMetadata } from '~/types/api'

interface Props {
  pagination: PaginationMetadata
  siblingCount?: number
}

const props = withDefaults(defineProps<Props>(), {
  siblingCount: 1
})

defineEmits<{
  'page-change': [page: number]
}>()

const visiblePages = computed(() => {
  const { page: currentPage, total_pages: totalPages } = props.pagination
  const { siblingCount } = props

  if (totalPages <= 5) {
    return Array.from({ length: totalPages }, (_, i) => i + 1)
  }

  const leftSiblingIndex = Math.max(currentPage - siblingCount, 1)
  const rightSiblingIndex = Math.min(currentPage + siblingCount, totalPages)

  const shouldShowLeftDots = leftSiblingIndex > 2
  const shouldShowRightDots = rightSiblingIndex < totalPages - 2

  const firstPageIndex = 1
  const lastPageIndex = totalPages

  if (!shouldShowLeftDots && shouldShowRightDots) {
    const leftItemCount = 3 + 2 * siblingCount
    const leftRange = Array.from({ length: leftItemCount }, (_, i) => i + 1)

    return [...leftRange, '...', totalPages]
  }

  if (shouldShowLeftDots && !shouldShowRightDots) {
    const rightItemCount = 3 + 2 * siblingCount
    const rightRange = Array.from(
      { length: rightItemCount },
      (_, i) => totalPages - rightItemCount + i + 1
    )

    return [firstPageIndex, '...', ...rightRange]
  }

  if (shouldShowLeftDots && shouldShowRightDots) {
    const middleRange = Array.from(
      { length: rightSiblingIndex - leftSiblingIndex + 1 },
      (_, i) => leftSiblingIndex + i
    )

    return [firstPageIndex, '...', ...middleRange, '...', lastPageIndex]
  }

  return []
})
</script>
