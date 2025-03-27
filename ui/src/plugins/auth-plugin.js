import { useAuthStore } from '@/stores/auth';

export default {
  install: (app) => {
    const authStore = useAuthStore();
    
    // Initialize auth state
    authStore.init();
    
    // Add global navigation guard for protected routes
    app.config.globalProperties.$authCheck = () => {
      return authStore.isAuthenticated;
    };
  }
};