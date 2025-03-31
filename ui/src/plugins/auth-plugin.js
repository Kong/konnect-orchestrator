import { useAuthStore } from '@/stores/auth';

export default {
  install: (app) => {
    // Access the store AFTER Pinia has been installed
    const authStore = useAuthStore();
    
    // Initialize auth state
    setTimeout(() => {
      // Use setTimeout to ensure this runs after Pinia is fully initialized
      authStore.init();
    }, 0);
    
    // Add global navigation guard for protected routes
    app.config.globalProperties.$authCheck = () => {
      return authStore.isAuthenticated;
    };
  }
};