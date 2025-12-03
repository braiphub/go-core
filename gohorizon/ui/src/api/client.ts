import axios from 'axios'
import type { Stats, FailedJob, RecentJob, Supervisor, Workload, MetricSnapshot } from '@/types'

// Get the base API path - works for both dev and embedded deployment
function getApiBasePath(): string {
  // In dev mode, we know it's /horizon/api
  if (import.meta.env.DEV) {
    return '/horizon/api'
  }
  // In production (embedded), the API is always at the same base path + /api
  // e.g., if dashboard is at /horizon/, API is at /horizon/api/
  const currentPath = window.location.pathname
  // Find the horizon base by removing any SPA route suffix
  const match = currentPath.match(/^(\/[^/]+)/)
  const basePath = match ? match[1] : '/horizon'
  return `${basePath}/api`
}

const api = axios.create({
  baseURL: getApiBasePath(),
  timeout: 10000,
})

export const horizonApi = {
  // Stats
  async getStats(): Promise<Stats> {
    const { data } = await api.get<Stats>('/stats')
    return data
  },

  // Queues
  async getQueues(): Promise<Stats['queues']> {
    const { data } = await api.get<Stats['queues']>('/queues')
    return data
  },

  // Workload
  async getWorkload(): Promise<Workload[]> {
    const { data } = await api.get<Workload[]>('/workload')
    return data
  },

  // Supervisors
  async getSupervisors(): Promise<Supervisor[]> {
    const { data } = await api.get<Supervisor[]>('/supervisors')
    return data
  },

  // Recent Jobs
  async getRecentJobs(limit = 50): Promise<RecentJob[]> {
    const { data } = await api.get<RecentJob[]>('/jobs/recent', { params: { limit } })
    return data
  },

  // Failed Jobs
  async getFailedJobs(limit = 50): Promise<FailedJob[]> {
    const { data } = await api.get<FailedJob[]>('/jobs/failed', { params: { limit } })
    return data
  },

  async retryJob(id: string): Promise<void> {
    await api.post('/jobs/retry', { id })
  },

  async retryAllJobs(): Promise<void> {
    await api.post('/jobs/retry-all')
  },

  async flushFailedJobs(): Promise<void> {
    await api.post('/jobs/flush')
  },

  // Metrics
  async getMetricSnapshots(): Promise<MetricSnapshot[]> {
    const { data } = await api.get<MetricSnapshot[]>('/metrics/snapshots')
    return data
  },
}

export default api
