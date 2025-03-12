<template>
    <div class="callback">
      <div v-if="loading" class="loading-container">
        <div class="spinner"></div>
        <p>Processing GitHub authentication...</p>
      </div>
      <div v-else-if="error" class="error-container">
        <h2>Authentication Error</h2>
        <p>{{ error }}</p>
        <router-link to="/" class="back-link">Back to Home</router-link>
      </div>
    </div>
  </template>
  
  <script setup>
  import { ref, onMounted } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useAuthStore } from '@/stores/auth';
  import axios from 'axios';
  import { setCsrfToken } from '@/services/api';
  
  const route = useRoute();
  const router = useRouter();
  const authStore = useAuthStore();
  
  // State
  const loading = ref(true);
  const error = ref(null);
  
  // Process the callback from GitHub
  onMounted(async () => {
    try {
      // Get the token from the URL
      const token = route.query.token;
      
      if (!token) {
        error.value = 'No authentication token received';
        loading.value = false;
        return;
      }
      
      // Make a request to the success endpoint to get CSRF token
      const response = await axios.get(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'}/auth/success`, {
        params: { token },
        withCredentials: true
      });
      
      if (response.data.csrf_token) {
        // Store the CSRF token
        setCsrfToken(response.data.csrf_token);
        
        // Store the auth token and load user info
        const success = authStore.handleCallback(token, response.data.csrf_token);
        
        if (success) {
          // Redirect to dashboard
          router.push('/dashboard');
        } else {
          error.value = 'Failed to process authentication';
          loading.value = false;
        }
      } else {
        error.value = 'CSRF token not provided by the server';
        loading.value = false;
      }
    } catch (err) {
      console.error('Callback error:', err);
      error.value = 'An error occurred during authentication';
      loading.value = false;
    }
  });
  </script>
  
  <style scoped>
  .callback {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 50vh;
    padding: 2rem;
  }
  
  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
  }
  
  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid rgba(0, 0, 0, 0.1);
    border-radius: 50%;
    border-top-color: #0366d6;
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
  }
  
  .error-container {
    text-align: center;
    max-width: 500px;
    padding: 2rem;
    background-color: #ffdce0;
    border: 1px solid #cb2431;
    border-radius: 6px;
  }
  
  .error-container h2 {
    color: #cb2431;
    margin-top: 0;
  }
  
  .back-link {
    display: inline-block;
    margin-top: 1rem;
    padding: 8px 16px;
    background-color: #0366d6;
    color: white;
    border-radius: 6px;
    text-decoration: none;
    font-weight: 500;
    transition: background-color 0.2s;
  }
  
  .back-link:hover {
    background-color: #0250a4;
  }
  
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
  </style>