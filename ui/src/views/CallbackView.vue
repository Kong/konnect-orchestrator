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
    <div v-else class="success-container">
      <p>Authentication successful! Redirecting to dashboard...</p>
      <button @click="manualRedirect" class="redirect-button">
        Click here if you're not redirected automatically
      </button>
    </div>
  </div>
</template>

<script setup>
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import api from '@/services/api';
import { useToast } from '@/composables/useToast';
import { onMounted, ref } from 'vue';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const toast = useToast();

// State
const loading = ref(true);
const error = ref(null);

// Manual redirect function as backup
const manualRedirect = () => {
  console.log('Manual redirect triggered');
  router.push('/dashboard');
};

// Process the callback from GitHub
onMounted(async () => {
  const code = route.query.code;
  
  if (!code) {
    loading.value = false;
    error.value = 'No authentication code provided';
    toast.error('Authentication failed: No code provided');
    return;
  }
  
  try {
    // Verify the authentication code
    const response = await api.auth.verifyCode(code);
    
    if (response.data.success) {
      // Set the CSRF token
      authStore.setSessionCsrfToken(response.data.csrf_token);
      
      // Load user data explicitly
      const userLoaded = await authStore.loadUser();
      
      if (userLoaded) {
        toast.success('Login successful!');
        
        // Set loading to false before redirect
        loading.value = false;
        
        // Small delay to ensure state is updated
        setTimeout(() => {
          const redirect = route.query.redirect || '/dashboard';
          router.push(redirect);
        }, 100);
      } else {
        loading.value = false;
        error.value = 'Failed to load user profile';
        toast.error('Authentication failed: Could not load user profile');
      }
    } else {
      loading.value = false;
      error.value = 'Authentication verification failed';
      toast.error('Authentication failed');
    }
  } catch (err) {
    console.error('Auth error:', err);
    loading.value = false;
    error.value = err.message || 'Unknown authentication error';
    toast.error('Authentication failed: ' + (err.message || 'Unknown error'));
  }
});
</script>

<style scoped>
/* Keep your existing styles and add: */
.success-container {
  text-align: center;
  padding: 2rem;
  background-color: #f0fff4;
  border: 1px solid #68d391;
  border-radius: 6px;
}

.redirect-button {
  margin-top: 1rem;
  padding: 8px 16px;
  background-color: #0366d6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
}

.redirect-button:hover {
  background-color: #0250a4;
}

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
  to {
    transform: rotate(360deg);
  }
}
</style>