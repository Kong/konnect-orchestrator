// Authentication store using Pinia
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import api, { setCsrfToken } from '@/services/api';

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref(null);
  const loading = ref(false);
  const error = ref(null);
  
  // Getters (computed)
  const isAuthenticated = computed(() => {
    return !!user.value && !!localStorage.getItem('auth_token');
  });
  
  const avatar = computed(() => {
    return user.value?.avatar_url || '';
  });
  
  const username = computed(() => {
    return user.value?.login || '';
  });
  
  // Actions
  async function loadUser() {
    if (!localStorage.getItem('auth_token')) return;
    
    loading.value = true;
    error.value = null;
    
    try {
      const response = await api.user.getProfile();
      user.value = response.data;
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to load user profile';
      // If authentication fails, clear the token
      if (err.response?.status === 401) {
        logout();
      }
    } finally {
      loading.value = false;
    }
  }
  
  function initiateLogin() {
    // Redirect to the GitHub login URL
    window.location.href = api.auth.getLoginUrl();
  }
  
  function handleCallback(token, csrfToken) {
    // Store the token and load user info
    if (token) {
      localStorage.setItem('auth_token', token);
      
      // Store CSRF token if provided
      if (csrfToken) {
        setCsrfToken(csrfToken);
      }
      
      loadUser();
      return true;
    }
    return false;
  }
  
  async function logout() {
    try {
      // Call logout endpoint if authenticated
      if (isAuthenticated.value) {
        await api.auth.logout();
      }
    } catch (err) {
      console.error('Logout error:', err);
    } finally {
      // Clear local state regardless of API success
      user.value = null;
      localStorage.removeItem('auth_token');
      setCsrfToken(''); // Clear CSRF token
    }
  }
  
  // Check token on initialization
  function checkAuth() {
    const token = localStorage.getItem('auth_token');
    if (token && !user.value) {
      loadUser();
    }
  }
  
  // Return state, getters, and actions
  return {
    // State
    user,
    loading,
    error,
    
    // Getters
    isAuthenticated,
    avatar,
    username,
    
    // Actions
    loadUser,
    initiateLogin,
    handleCallback,
    logout,
    checkAuth
  };
});