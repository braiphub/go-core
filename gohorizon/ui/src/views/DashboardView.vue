<script setup lang="ts">
import { computed } from 'vue'
import { horizonApi } from '@/api/client'
import { usePolling } from '@/composables/usePolling'
import StatCard from '@/components/StatCard.vue'
import QueueCard from '@/components/QueueCard.vue'

const { data: stats, loading, error } = usePolling(() => horizonApi.getStats(), 3000)

const statusColor = computed(() => {
  if (!stats.value) return 'gray'
  switch (stats.value.status) {
    case 'running': return 'green'
    case 'paused': return 'yellow'
    default: return 'red'
  }
})

const formatNumber = (n: number) => {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return n.toString()
}
</script>

<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
      <p class="text-gray-500">Queue system overview and metrics</p>
    </div>

    <!-- Loading state -->
    <div v-if="loading && !stats" class="flex items-center justify-center h-64">
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

    <!-- Stats -->
    <div v-else-if="stats">
      <!-- Status banner -->
      <div class="mb-6 p-4 rounded-lg flex items-center justify-between"
           :class="{
             'bg-green-50 border border-green-200': stats.status === 'running',
             'bg-yellow-50 border border-yellow-200': stats.status === 'paused',
             'bg-red-50 border border-red-200': stats.status === 'stopped',
           }">
        <div class="flex items-center">
          <span class="w-3 h-3 rounded-full mr-3"
                :class="{
                  'bg-green-500 animate-pulse': stats.status === 'running',
                  'bg-yellow-500': stats.status === 'paused',
                  'bg-red-500': stats.status === 'stopped',
                }"></span>
          <span class="font-medium capitalize"
                :class="{
                  'text-green-800': stats.status === 'running',
                  'text-yellow-800': stats.status === 'paused',
                  'text-red-800': stats.status === 'stopped',
                }">
            {{ stats.status }}
          </span>
        </div>
        <span class="text-sm text-gray-500">
          Last updated: {{ new Date(stats.updated_at).toLocaleTimeString() }}
        </span>
      </div>

      <!-- Main stats -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard
          title="Jobs Per Minute"
          :value="stats.jobs_per_minute.toFixed(1)"
          subtitle="Current throughput"
          color="blue"
        />
        <StatCard
          title="Total Processed"
          :value="formatNumber(stats.total_processed)"
          subtitle="All time"
          color="green"
        />
        <StatCard
          title="Total Failed"
          :value="formatNumber(stats.total_failed)"
          subtitle="Requires attention"
          :color="stats.total_failed > 0 ? 'red' : 'gray'"
        />
        <StatCard
          title="Active Workers"
          :value="stats.total_workers"
          subtitle="Processing jobs"
          color="blue"
        />
      </div>

      <!-- Secondary stats -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
        <StatCard
          title="Pending Jobs"
          :value="formatNumber(stats.total_pending)"
          subtitle="Waiting to be processed"
          :color="stats.total_pending > 100 ? 'yellow' : 'gray'"
        />
        <StatCard
          title="Queues"
          :value="stats.queues.length"
          subtitle="Active queues"
          color="gray"
        />
        <StatCard
          title="Fail Rate"
          :value="stats.total_processed > 0 ? ((stats.total_failed / stats.total_processed) * 100).toFixed(2) + '%' : '0%'"
          subtitle="Overall failure rate"
          :color="stats.total_failed / stats.total_processed > 0.05 ? 'red' : 'green'"
        />
      </div>

      <!-- Queues -->
      <div class="mb-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Queues</h2>
        <div v-if="stats.queues.length === 0" class="card p-8 text-center">
          <p class="text-gray-500">No queues configured</p>
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <QueueCard v-for="queue in stats.queues" :key="queue.queue" :queue="queue" />
        </div>
      </div>
    </div>
  </div>
</template>
