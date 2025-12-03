export interface QueueStats {
  queue: string
  total_processed: number
  total_failed: number
  pending_jobs: number
  reserved_jobs: number
  delayed_jobs: number
  jobs_per_minute: number
  fail_rate: number
}

export interface Stats {
  status: 'running' | 'paused' | 'stopped'
  jobs_per_minute: number
  total_processed: number
  total_failed: number
  total_pending: number
  total_workers: number
  queues: QueueStats[]
  updated_at: string
}

export interface FailedJob {
  id: string
  queue: string
  payload: JobPayload
  exception: string
  failed_at: string
}

export interface JobPayload {
  id: string
  name: string
  queue: string
  data: Record<string, unknown>
  attempts: number
  max_attempts: number
  tags?: string[]
  created_at: string
  available_at: string
  reserved_at?: string
  timeout: number
  retry_delay: number
}

export interface RecentJob {
  id: string
  name: string
  queue: string
  status: 'pending' | 'reserved' | 'completed' | 'failed'
  attempts: number
  runtime: number
  completed_at: string
  tags?: string[]
}

export interface Supervisor {
  name: string
  status: 'running' | 'paused'
  queues: string[]
  workers: number
  min_workers: number
  max_workers: number
  balance_mode: string
}

export interface Workload {
  queue: string
  length: number
  wait: number
  processes: number
}

export interface MetricSnapshot {
  timestamp: string
  jobs_per_minute: number
  total_processed: number
  total_failed: number
  queues: Record<string, number>
}
