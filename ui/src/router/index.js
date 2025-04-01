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
  console.log('Navigation guard running', { 
    to: to.path,
    from: from.path
  });
  
  // Update document title
  document.title = to.meta.title || 'GitHub Explorer';
  
  const authStore = useAuthStore();
  
  // Skip auth checks if we recently logged out
  if (authStore.recentlyLoggedOut) {
    console.log('Recently logged out, skipping auth checks');
    authStore.recentlyLoggedOut = false;
    return next();
  }
  
  // Skip initialization if authentication has already failed
  // This prevents repeated API calls to /api/user when logged out
  const needsInit = !authStore.initialized && 
                   to.matched.some(record => record.meta.requiresAuth) &&
                   !authStore.authenticationFailed;
                   
  // Only initialize if necessary
  if (needsInit) {
    await authStore.init();
  }
  
  // If auth is loading, wait for it to complete
  if (authStore.loading.value) {
    console.log('Auth is loading, waiting...');
    await new Promise(resolve => setTimeout(resolve, 200));
  }
  
  // Check if the route requires authentication
  if (to.matched.some(record => record.meta.requiresAuth)) {
    console.log('Route requires auth, checking authentication state...');
    console.log('Auth state:', { 
      initialized: authStore.initialized,
      isAuthenticated: authStore.isAuthenticated,
      authFailed: authStore.authenticationFailed
    });
    
    // If we're not authenticated and trying to access a protected route
    if (!authStore.isAuthenticated) {
      console.log('Not authenticated, redirecting to home');
      next({
        path: '/',
        query: { redirect: to.fullPath }
      });
    } else {
      console.log('Authentication verified, proceeding to route');
      next();
    }
  } else {
    // No auth required, proceed
    next();
  }
});

export default router;