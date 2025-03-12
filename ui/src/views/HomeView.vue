<template>
    <div class="home">
      <div class="hero">
        <h1>GitHub Repository Explorer</h1>
        <p>Connect with GitHub to browse your repositories</p>
        
        <GitHubLogin />
      </div>
      
      <div class="features">
        <div class="feature">
          <h3>Browse Repositories</h3>
          <p>View all your GitHub repositories in one place</p>
        </div>
        
        <div class="feature">
          <h3>Repository Details</h3>
          <p>Quick access to repository information</p>
        </div>
        
        <div class="feature">
          <h3>Private Access</h3>
          <p>Securely access your private repositories</p>
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import GitHubLogin from '@/components/GitHubLogin.vue';
  import { useRouter } from 'vue-router';
  import { useAuthStore } from '@/stores/auth';
  import { onMounted, watch } from 'vue';
  
  const router = useRouter();
  const authStore = useAuthStore();
  
  // Redirect to dashboard if already authenticated
  const checkAuth = () => {
    if (authStore.isAuthenticated) {
      router.push('/dashboard');
    }
  };
  
  onMounted(() => {
    checkAuth();
  });
  
  // Watch for authentication state changes
  watch(() => authStore.isAuthenticated, (isAuthenticated) => {
    if (isAuthenticated) {
      router.push('/dashboard');
    }
  });
  </script>
  
  <style scoped>
  .home {
    max-width: 800px;
    margin: 0 auto;
    padding: 2rem;
  }
  
  .hero {
    text-align: center;
    margin-bottom: 3rem;
  }
  
  h1 {
    font-size: 2rem;
    margin-bottom: 0.5rem;
    color: #24292e;
  }
  
  p {
    color: #586069;
    margin-bottom: 1.5rem;
  }
  
  .features {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 2rem;
    margin-top: 3rem;
  }
  
  .feature {
    background-color: #f6f8fa;
    border: 1px solid #e1e4e8;
    border-radius: 6px;
    padding: 1.5rem;
    text-align: center;
  }
  
  .feature h3 {
    font-size: 1.2rem;
    margin-bottom: 0.75rem;
    color: #24292e;
  }
  
  .feature p {
    font-size: 0.9rem;
    color: #586069;
  }
  </style>