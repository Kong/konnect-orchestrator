// Authentication store using Pinia
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import api, { setCsrfToken as setApiCsrfToken } from '@/services/api';
import router from '@/router';

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref(null);
  const loading = ref(false);
  const error = ref(null);
  const initialized = ref(false);
  const intentionallyLoggedOut = ref(false);
  
  // Getters (computed)
  const isAuthenticated = computed(() => {
    return !!user.value;
  });
  
  const avatar = computed(() => {
    return user.value?.avatar_url || '';
  });
  
  const username = computed(() => {
    return user.value?.login || '';
  });
  
  // Actions
  async function loadUser() {
    if (intentionallyLoggedOut.value) {
      return false;
    }
    loading.value = true;
    error.value = null;
    
    try {
      console.log('Loading user profile...');
      const response = await api.user.getProfile();
      
      // If we successfully got the user profile, we're authenticated
      user.value = response.data;
      
      // Set initialized flag
      initialized.value = true;
      
      console.log('User loaded successfully:', !!user.value);
      return true;
    } catch (err) {
      console.error('Failed to load user profile:', err);
      error.value = err.response?.data?.error || 'Failed to load user profile';
      
      // Clear user on authentication errors
      if (err.response?.status === 401) {
        user.value = null;
      }
      
      return false;
    } finally {
      loading.value = false;
    }
  }

  function setCsrfToken(token) {
    // Update the token in the API client
    setApiCsrfToken(token);
  }

  function initiateLogin() {
    // Redirect to the GitHub login URL
    window.location.href = api.auth.getLoginUrl();
  }
  
  function setSessionCsrfToken(csrfToken) {
    if (csrfToken) {
      setCsrfToken(csrfToken);
      return true;
    }
    return false;
  }
  
  async function logout() {
    try {
      if (isAuthenticated.value) {
        await api.auth.logout();
      }
    } catch (err) {
      console.error('Logout error:', err);
    } finally {
      user.value = null;
      setCsrfToken('');
      localStorage.removeItem('auth_token');
      intentionallyLoggedOut.value = true; // Set flag to prevent further checks
      
      if (router && router.currentRoute && router.currentRoute.value.path !== '/') {
        router.push('/');
      }
    }
  }
  
  function init() {
    if (initialized.value) return Promise.resolve();
    
    return new Promise(async (resolve) => {
      try {
        // Add a check here to prevent unnecessary API calls
        if (!localStorage.getItem('auth_token')) {
          initialized.value = true;
          return resolve();
        }
        
        // Just try to load user - this will use the cookie if present
        await loadUser();
      } catch (error) {
        console.error('Failed to load user during initialization:', error);
      }
      initialized.value = true;
      resolve();
    });
  }
  // Return state, getters, and actions
  return {
    // State
    user,
    loading,
    error,
    initialized,
    
    // Getters
    isAuthenticated,
    avatar,
    username,
    
    // Actions
    loadUser,
    setCsrfToken,
    initiateLogin,
    setSessionCsrfToken,
    logout,
    init
  };
});