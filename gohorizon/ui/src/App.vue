<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'

const route = useRoute()
const sidebarOpen = ref(false)

const navigation = [
  { name: 'Dashboard', href: '/', icon: 'chart-bar' },
  { name: 'Queues', href: '/queues', icon: 'queue-list' },
  { name: 'Recent Jobs', href: '/jobs/recent', icon: 'clock' },
  { name: 'Failed Jobs', href: '/jobs/failed', icon: 'exclamation-triangle' },
  { name: 'Supervisors', href: '/supervisors', icon: 'server' },
]

const isActive = (href: string) => {
  if (href === '/') return route.path === '/'
  return route.path.startsWith(href)
}
</script>

<template>
  <div class="min-h-screen bg-gray-100">
    <!-- Sidebar -->
    <div class="fixed inset-y-0 left-0 z-50 w-64 bg-horizon-900 transform transition-transform duration-200 ease-in-out lg:translate-x-0"
         :class="{ '-translate-x-full': !sidebarOpen, 'translate-x-0': sidebarOpen }">
      <!-- Logo -->
      <div class="flex items-center justify-center h-16 bg-horizon-950">
        <h1 class="text-xl font-bold text-white flex items-center gap-2">
          <svg class="w-8 h-8" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/>
            <path d="M12 6v6l4 2" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          Horizon
        </h1>
      </div>

      <!-- Navigation -->
      <nav class="mt-8 px-4 space-y-1">
        <RouterLink
          v-for="item in navigation"
          :key="item.name"
          :to="item.href"
          class="flex items-center px-4 py-3 text-sm font-medium rounded-lg transition-colors"
          :class="isActive(item.href)
            ? 'bg-horizon-800 text-white'
            : 'text-horizon-300 hover:bg-horizon-800 hover:text-white'"
          @click="sidebarOpen = false"
        >
          <!-- Icons -->
          <svg v-if="item.icon === 'chart-bar'" class="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
          </svg>
          <svg v-else-if="item.icon === 'queue-list'" class="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16"/>
          </svg>
          <svg v-else-if="item.icon === 'clock'" class="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          <svg v-else-if="item.icon === 'exclamation-triangle'" class="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
          </svg>
          <svg v-else-if="item.icon === 'server'" class="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
          </svg>
          {{ item.name }}
        </RouterLink>
      </nav>

      <!-- Status indicator -->
      <div class="absolute bottom-0 left-0 right-0 p-4 border-t border-horizon-800">
        <div class="flex items-center text-horizon-400 text-sm">
          <span class="w-2 h-2 bg-green-500 rounded-full mr-2 animate-pulse"></span>
          Connected
        </div>
      </div>
    </div>

    <!-- Mobile sidebar toggle -->
    <div class="lg:hidden fixed top-0 left-0 right-0 z-40 bg-white shadow-sm">
      <div class="flex items-center justify-between h-16 px-4">
        <button @click="sidebarOpen = !sidebarOpen" class="text-gray-500 hover:text-gray-700">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>
          </svg>
        </button>
        <h1 class="text-lg font-semibold text-gray-900">Horizon</h1>
        <div class="w-6"></div>
      </div>
    </div>

    <!-- Main content -->
    <div class="lg:pl-64">
      <main class="py-6 px-4 sm:px-6 lg:px-8 lg:pt-6 pt-20">
        <RouterView />
      </main>
    </div>

    <!-- Mobile overlay -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 bg-black bg-opacity-50 z-40 lg:hidden"
      @click="sidebarOpen = false"
    ></div>
  </div>
</template>
