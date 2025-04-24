// src/plugins/auth-plugin.js
import { useAuthStore } from '@/stores/auth';

export default {
  install: (app) => {
    const authStore = useAuthStore();
    
    // Don't check localStorage, just initialize if we haven't failed auth already
    if (!authStore.authenticationFailed) {
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