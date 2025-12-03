import { ref, onMounted, onUnmounted, type Ref } from 'vue'

export function usePolling<T>(
  fetchFn: () => Promise<T>,
  interval = 3000
): {
  data: Ref<T | null>
  loading: Ref<boolean>
  error: Ref<Error | null>
  refresh: () => Promise<void>
} {
  const data = ref<T | null>(null) as Ref<T | null>
  const loading = ref(true)
  const error = ref<Error | null>(null)
  let timer: ReturnType<typeof setInterval> | null = null

  const fetch = async () => {
    try {
      data.value = await fetchFn()
      error.value = null
    } catch (e) {
      error.value = e as Error
    } finally {
      loading.value = false
    }
  }

  const start = () => {
    fetch()
    timer = setInterval(fetch, interval)
  }

  const stop = () => {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  onMounted(start)
  onUnmounted(stop)

  return {
    data,
    loading,
    error,
    refresh: fetch,
  }
}
