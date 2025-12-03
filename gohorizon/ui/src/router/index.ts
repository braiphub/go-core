import { createRouter, createWebHashHistory } from 'vue-router'
import DashboardView from '@/views/DashboardView.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: DashboardView,
    },
    {
      path: '/queues',
      name: 'queues',
      component: () => import('@/views/QueuesView.vue'),
    },
    {
      path: '/jobs/recent',
      name: 'recent-jobs',
      component: () => import('@/views/RecentJobsView.vue'),
    },
    {
      path: '/jobs/failed',
      name: 'failed-jobs',
      component: () => import('@/views/FailedJobsView.vue'),
    },
    {
      path: '/supervisors',
      name: 'supervisors',
      component: () => import('@/views/SupervisorsView.vue'),
    },
  ],
})

export default router
