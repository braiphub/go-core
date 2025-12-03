<script setup lang="ts">
defineProps<{
  title: string
  value: string | number
  subtitle?: string
  trend?: 'up' | 'down' | 'neutral'
  color?: 'blue' | 'green' | 'red' | 'yellow' | 'gray'
}>()
</script>

<template>
  <div class="stat-card">
    <div class="flex items-center justify-between">
      <h3 class="text-sm font-medium text-gray-500">{{ title }}</h3>
      <div v-if="trend" class="flex items-center">
        <svg v-if="trend === 'up'" class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18"/>
        </svg>
        <svg v-else-if="trend === 'down'" class="w-4 h-4 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3"/>
        </svg>
      </div>
    </div>
    <p class="mt-2 text-3xl font-bold"
       :class="{
         'text-horizon-600': color === 'blue' || !color,
         'text-green-600': color === 'green',
         'text-red-600': color === 'red',
         'text-yellow-600': color === 'yellow',
         'text-gray-600': color === 'gray',
       }">
      {{ value }}
    </p>
    <p v-if="subtitle" class="mt-1 text-sm text-gray-500">{{ subtitle }}</p>
  </div>
</template>
