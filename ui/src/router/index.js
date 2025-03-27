import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import HomeView from '@/views/HomeView.vue';
import DashboardView from '@/views/DashboardView.vue';
import CallbackView from '@/views/CallbackView.vue';

const routes = [
  {
    path: '/',
    name: 'home',
    component: HomeView,
    meta: {
      title: 'Home - GitHub Explorer'
    }
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: DashboardView,
    meta: {
      title: 'Dashboard - GitHub Explorer',
      requiresAuth: true
    }
  },
  {
    path: '/auth/success',
    name: 'auth-callback',
    component: CallbackView,
    meta: {
      title: 'Authentication Callback'
    }
  },
  {
    // Catch all route (404)
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
});

router.beforeEach(async (to, from, next) => {
  document.title = to.meta.title || 'GitHub Explorer';
  
  const authStore = useAuthStore();
  
  // Wait for auth to initialize
  await authStore.init();
  
  // Check if route requires authentication
  if (to.matched.some(record => record.meta.requiresAuth)) {
    if (!authStore.isAuthenticated) {
      next({
        path: '/',
        query: { redirect: to.fullPath }
      });
    } else {
      next();
    }
  } else {
    next();
  }
});

export default router;