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
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

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
  try {
    const code = route.query.code;
    
    if (!code) {
      error.value = 'No authentication code received';
      loading.value = false;
      return;
    }

    // Use the api service method instead of direct axios call
    const response = await api.auth.verifyCode(code);
    
    if (response.data.csrf_token) {
      // Store CSRF token using the auth store method
      authStore.setCsrfToken(response.data.csrf_token);
      
      // Load user profile
      await authStore.loadUser();
      
      // Navigate to dashboard
      router.push('/dashboard');
    } else {
      error.value = 'Authentication failed - no CSRF token received';
    }
  } catch (err) {
    console.error('Authentication error:', err);
    error.value = 'Authentication failed. Please try again.';
  } finally {
    loading.value = false;
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