import { useAuthStore } from '@/stores/auth';

export default {
  install: (app) => {
    const authStore = useAuthStore();
    
    // Check local storage for auth token to prevent unnecessary API calls
    const hasToken = localStorage.getItem('auth_token') || 
                   document.cookie.includes('auth_token=');
    
    // Only try to initialize if we have a token or haven't tried before
    if (hasToken && !authStore.authenticationFailed) {
      authStore.init().catch(err => {
        console.error('Auth initialization error:', err);
      });
    } else {
      // Mark as initialized without making API calls
      authStore.initialized = true;
    }
    
    // Add global nav guard for protected routes
    app.config.globalProperties.$authCheck = () => {
      return authStore.isAuthenticated;
    };
  }
};