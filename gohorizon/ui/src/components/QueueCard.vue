<script setup lang="ts">
import type { QueueStats } from '@/types'

defineProps<{
  queue: QueueStats
}>()

const formatNumber = (n: number) => {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return n.toString()
}

const formatRate = (rate: number) => {
  return (rate * 100).toFixed(2) + '%'
}
</script>

<template>
  <div class="card p-4 hover:shadow-md transition-shadow">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-semibold text-gray-900">{{ queue.queue }}</h3>
      <span class="px-2 py-1 text-xs font-medium rounded-full"
            :class="queue.pending_jobs > 0 ? 'bg-yellow-100 text-yellow-800' : 'bg-green-100 text-green-800'">
        {{ queue.pending_jobs > 0 ? 'Active' : 'Idle' }}
      </span>
    </div>

    <div class="grid grid-cols-3 gap-4 text-center">
      <div>
        <p class="text-2xl font-bold text-horizon-600">{{ formatNumber(queue.pending_jobs) }}</p>
        <p class="text-xs text-gray-500">Pending</p>
      </div>
      <div>
        <p class="text-2xl font-bold text-green-600">{{ formatNumber(queue.total_processed) }}</p>
        <p class="text-xs text-gray-500">Processed</p>
      </div>
      <div>
        <p class="text-2xl font-bold text-red-600">{{ formatNumber(queue.total_failed) }}</p>
        <p class="text-xs text-gray-500">Failed</p>
      </div>
    </div>

    <div class="mt-4 pt-4 border-t border-gray-100">
      <div class="flex justify-between text-sm">
        <span class="text-gray-500">Throughput</span>
        <span class="font-medium">{{ queue.jobs_per_minute.toFixed(1) }}/min</span>
      </div>
      <div class="flex justify-between text-sm mt-1">
        <span class="text-gray-500">Fail Rate</span>
        <span class="font-medium" :class="queue.fail_rate > 0.05 ? 'text-red-600' : 'text-gray-900'">
          {{ formatRate(queue.fail_rate) }}
        </span>
      </div>
      <div class="flex justify-between text-sm mt-1">
        <span class="text-gray-500">Reserved</span>
        <span class="font-medium">{{ queue.reserved_jobs }}</span>
      </div>
      <div class="flex justify-between text-sm mt-1">
        <span class="text-gray-500">Delayed</span>
        <span class="font-medium">{{ queue.delayed_jobs }}</span>
      </div>
    </div>
  </div>
</template>
