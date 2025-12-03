<script setup lang="ts">
import { horizonApi } from '@/api/client'
import { usePolling } from '@/composables/usePolling'

const { data: supervisors, loading, error } = usePolling(() => horizonApi.getSupervisors(), 5000)
</script>

<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Supervisors</h1>
      <p class="text-gray-500">Worker pools and their status</p>
    </div>

    <!-- Loading state -->
    <div v-if="loading && !supervisors" class="flex items-center justify-center h-64">
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

    <!-- Supervisors grid -->
    <div v-else-if="supervisors">
      <div v-if="supervisors.length === 0" class="card p-8 text-center">
        <svg class="w-12 h-12 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
        </svg>
        <h3 class="text-lg font-medium text-gray-900 mb-2">No Supervisors</h3>
        <p class="text-gray-500">No supervisors have been configured yet.</p>
      </div>
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div v-for="supervisor in supervisors" :key="supervisor.name" class="card p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-semibold text-gray-900">{{ supervisor.name }}</h3>
            <span class="px-2 py-1 text-xs font-medium rounded-full capitalize"
                  :class="supervisor.status === 'running' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'">
              {{ supervisor.status }}
            </span>
          </div>

          <!-- Workers count -->
          <div class="mb-4">
            <div class="flex justify-between text-sm mb-1">
              <span class="text-gray-500">Active Workers</span>
              <span class="font-medium text-2xl text-horizon-600">{{ supervisor.workers ?? 0 }}</span>
            </div>
          </div>

          <!-- Details -->
          <div class="space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="text-gray-500">Balance Mode</span>
              <span class="font-medium capitalize">{{ supervisor.balance || 'auto' }}</span>
            </div>
          </div>

          <!-- Queues -->
          <div class="mt-4 pt-4 border-t border-gray-100">
            <p class="text-sm text-gray-500 mb-2">Queues</p>
            <div class="flex flex-wrap gap-2">
              <span v-for="queue in supervisor.queues" :key="queue"
                    class="px-2 py-1 text-xs bg-gray-100 text-gray-700 rounded">
                {{ queue }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
