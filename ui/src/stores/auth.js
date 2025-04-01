// Authentication store using Pinia
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import api, { setCsrfToken as setApiCsrfToken, getCsrfToken } from '@/services/api';

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref(null);
  const loading = ref(false);
  const error = ref(null);
  const initialized = ref(false);
  const recentlyLoggedOut = ref(false);
  
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
      const currentCSRFToken = getCsrfToken();
      console.log('CSRF token for logout:', currentCSRFToken);
  
      if (isAuthenticated.value) {
        try {
          // Try server-side logout with the CSRF token
          if (currentCSRFToken) {
            await api.auth.logout({
              headers: {
                'X-CSRF-Token': currentCSRFToken
              }
            });
          } else {
            // If no CSRF token, try without it as a fallback
            await api.auth.logout();
          }
        } catch (err) {
          console.error('Server logout error:', err);
          // Continue with client-side logout anyway
        }
        
        // Clear auth cookie manually since server logout might have failed
        document.cookie = "auth_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
      }
    } catch (err) {
      console.error('Logout error:', err);
    } finally {
      // Clear all client-side state
      user.value = null;
      initialized.value = false;  // Force re-initialization on next check
      setCsrfToken('');
      localStorage.removeItem('auth_token');
      
      // Set flag to prevent immediate re-authentication
      recentlyLoggedOut.value = true;
    }
  }
  
  function init() {
    if (initialized.value && user.value) return Promise.resolve(true);
    
    return new Promise(async (resolve) => {
      try {
        // Try to load user - this will use the cookie if present
        const success = await loadUser();
        return resolve(success);
      } catch (error) {
        console.error('Failed to load user during initialization:', error);
        return resolve(false);
      } finally {
        initialized.value = true;
      }
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