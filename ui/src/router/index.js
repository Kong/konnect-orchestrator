import { createRouter, createWebHistory } from 'vue-router';
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

// Navigation guard to check authentication for protected routes
router.beforeEach((to, from, next) => {
  // Update document title
  document.title = to.meta.title || 'GitHub Explorer';
  
  // Check if route requires authentication
  if (to.matched.some(record => record.meta.requiresAuth)) {
    // Check if user is authenticated
    const token = localStorage.getItem('auth_token');
    if (!token) {
      // Redirect to login page
      next({
        path: '/',
        query: { redirect: to.fullPath }
      });
    } else {
      // Continue to route
      next();
    }
  } else {
    // Continue to route
    next();
  }
});

export default router;