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
  const authenticationFailed = ref(false); // Add this flag
  
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
    // Prevent unnecessary API calls if we've already tried and failed
    if (authenticationFailed.value) {
      return false;
    }
    
    loading.value = true;
    error.value = null;
    
    try {
      console.log('Loading user profile...');
      const response = await api.user.getProfile();
      
      // If we successfully got the user profile, we're authenticated
      user.value = response.data;
      authenticationFailed.value = false; // Reset on success
      
      console.log('User loaded successfully:', !!user.value);
      return true;
    } catch (err) {
      console.error('Failed to load user profile:', err);
      error.value = err.response?.data?.error || 'Failed to load user profile';
      
      // Set the flag when authentication fails with 401
      if (err.response?.status === 401) {
        authenticationFailed.value = true;
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
    // Reset authentication failed flag when initiating login
    authenticationFailed.value = false;
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
      sessionStorage.removeItem('csrf_token'); // Also clear from sessionStorage
      
      // Set flags to prevent immediate re-authentication
      recentlyLoggedOut.value = true;
      authenticationFailed.value = true; // Also set this flag to prevent API thrashing
    }
  }
  
  function init() {
    // Skip initialization if authentication has already failed
    if (authenticationFailed.value) {
      initialized.value = true;
      return Promise.resolve(false);
    }
    
    // If already initialized and authenticated, return immediately
    if (initialized.value && user.value) {
      return Promise.resolve(true);
    }
    
    // Check for auth token before making API call
    const hasToken = localStorage.getItem('auth_token') || 
                    document.cookie.includes('auth_token=');
    
    if (!hasToken) {
      // No token means no auth, avoid API call
      initialized.value = true;
      return Promise.resolve(false);
    }
    
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
  
  // Reset all auth state (useful for testing or error recovery)
  function reset() {
    user.value = null;
    error.value = null;
    authenticationFailed.value = false;
    recentlyLoggedOut.value = false;
    initialized.value = false;
  }
  
  // Return state, getters, and actions
  return {
    // State
    user,
    loading,
    error,
    initialized,
    recentlyLoggedOut,
    authenticationFailed, // Include the new flag
    
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
    init,
    reset // Include the new method
  };
});