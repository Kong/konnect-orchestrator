import { useAuthStore } from '@/stores/auth';

export default {
  install: (app) => {
    const authStore = useAuthStore();
    
    // Initialize auth state, but check localStorage first
    if (localStorage.getItem('auth_token')) {
      authStore.init();
    } else {
      authStore.initialized = true; // Mark as initialized without API call
    }
    
    // Add global navigation guard for protected routes
    app.config.globalProperties.$authCheck = () => {
      return authStore.isAuthenticated;
    };
  }
};