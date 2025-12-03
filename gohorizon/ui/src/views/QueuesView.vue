<script setup lang="ts">
import { horizonApi } from '@/api/client'
import { usePolling } from '@/composables/usePolling'
import QueueCard from '@/components/QueueCard.vue'

const { data: queues, loading, error } = usePolling(() => horizonApi.getQueues(), 3000)
</script>

<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Queues</h1>
      <p class="text-gray-500">Monitor all queue metrics and performance</p>
    </div>

    <!-- Loading state -->
    <div v-if="loading && !queues" class="flex items-center justify-center h-64">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-horizon-600"></div>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="card p-6 text-center">
      <svg class="w-12 h-12 mx-auto text-red-500 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
      </svg>
      <h3 class="text-lg font-medium text-gray-900 mb-2">Connection Error</h3>
      <p class="text-gray-500">{{ error.message }}</p>
    </div>

    <!-- Queues grid -->
    <div v-else-if="queues">
      <div v-if="queues.length === 0" class="card p-8 text-center">
        <svg class="w-12 h-12 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16"/>
        </svg>
        <h3 class="text-lg font-medium text-gray-900 mb-2">No Queues</h3>
        <p class="text-gray-500">No queues have been configured yet.</p>
      </div>
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <QueueCard v-for="queue in queues" :key="queue.queue" :queue="queue" />
      </div>
    </div>
  </div>
</template>
