<script setup lang="ts">
import { ref } from 'vue'
import { horizonApi } from '@/api/client'
import { usePolling } from '@/composables/usePolling'

const { data: jobs, loading, error, refresh } = usePolling(() => horizonApi.getFailedJobs(100), 5000)

const retrying = ref<Set<string>>(new Set())
const retryingAll = ref(false)
const flushing = ref(false)
const expandedJob = ref<string | null>(null)

const formatTime = (dateStr: string) => {
  const date = new Date(dateStr)
  return date.toLocaleString()
}

const retryJob = async (id: string) => {
  retrying.value.add(id)
  try {
    await horizonApi.retryJob(id)
    await refresh()
  } finally {
    retrying.value.delete(id)
  }
}

const retryAll = async () => {
  retryingAll.value = true
  try {
    await horizonApi.retryAllJobs()
    await refresh()
  } finally {
    retryingAll.value = false
  }
}

const flushAll = async () => {
  if (!confirm('Are you sure you want to delete all failed jobs? This cannot be undone.')) {
    return
  }
  flushing.value = true
  try {
    await horizonApi.flushFailedJobs()
    await refresh()
  } finally {
    flushing.value = false
  }
}

const toggleExpand = (id: string) => {
  expandedJob.value = expandedJob.value === id ? null : id
}
</script>

<template>
  <div>
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Failed Jobs</h1>
        <p class="text-gray-500">Jobs that have failed and need attention</p>
      </div>
      <div v-if="jobs && jobs.length > 0" class="flex gap-2">
        <button
          @click="retryAll"
          :disabled="retryingAll"
          class="btn btn-primary"
        >
          <svg v-if="retryingAll" class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          Retry All
        </button>
        <button
          @click="flushAll"
          :disabled="flushing"
          class="btn btn-danger"
        >
          <svg v-if="flushing" class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          Flush All
        </button>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="loading && !jobs" class="flex items-center justify-center h-64">
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

    <!-- Jobs list -->
    <div v-else-if="jobs">
      <div v-if="jobs.length === 0" class="card p-8 text-center">
        <svg class="w-12 h-12 mx-auto text-green-500 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        <h3 class="text-lg font-medium text-gray-900 mb-2">No Failed Jobs</h3>
        <p class="text-gray-500">All jobs are processing successfully.</p>
      </div>
      <div v-else class="space-y-4">
        <div v-for="job in jobs" :key="job.id" class="card overflow-hidden">
          <div class="p-4">
            <div class="flex items-center justify-between">
              <div class="flex-1">
                <div class="flex items-center gap-3">
                  <h3 class="text-lg font-medium text-gray-900">{{ job.payload.name }}</h3>
                  <span class="px-2 py-1 text-xs font-medium bg-gray-100 text-gray-800 rounded">
                    {{ job.queue }}
                  </span>
                </div>
                <p class="text-sm text-gray-500 mt-1">
                  Failed at {{ formatTime(job.failed_at) }} • Attempt {{ job.payload.attempts }} of {{ job.payload.max_attempts }}
                </p>
              </div>
              <div class="flex items-center gap-2">
                <button
                  @click="toggleExpand(job.id)"
                  class="btn btn-secondary"
                >
                  {{ expandedJob === job.id ? 'Hide' : 'Details' }}
                </button>
                <button
                  @click="retryJob(job.id)"
                  :disabled="retrying.has(job.id)"
                  class="btn btn-success"
                >
                  <svg v-if="retrying.has(job.id)" class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Retry
                </button>
              </div>
            </div>

            <!-- Error message -->
            <div class="mt-3 p-3 bg-red-50 border border-red-200 rounded-lg">
              <p class="text-sm text-red-800 font-mono whitespace-pre-wrap">{{ job.exception }}</p>
            </div>

            <!-- Expanded details -->
            <div v-if="expandedJob === job.id" class="mt-4 pt-4 border-t border-gray-200">
              <h4 class="text-sm font-medium text-gray-700 mb-2">Job Payload</h4>
              <pre class="p-3 bg-gray-50 rounded-lg text-xs overflow-x-auto">{{ JSON.stringify(job.payload, null, 2) }}</pre>

              <div v-if="job.payload.tags && job.payload.tags.length > 0" class="mt-4">
                <h4 class="text-sm font-medium text-gray-700 mb-2">Tags</h4>
                <div class="flex flex-wrap gap-2">
                  <span v-for="tag in job.payload.tags" :key="tag"
                        class="px-2 py-1 text-xs bg-horizon-100 text-horizon-800 rounded">
                    {{ tag }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
