<template>
  <div class="home">
    <div class="hero">
      <h1>{{ title }}</h1>
      <h2 class="gradient-subtitle">Enabling <span class="gradient-text">self-service</span> API platforms.</h2>
      
      <div v-if="authStore.isAuthenticated" class="welcome-back">
        <p>Welcome back, {{ authStore.username }}!</p>
        <router-link to="/dashboard" class="cta-button">
          Go to Dashboard
        </router-link>
      </div>
    </div>
    
    <div class="features">
      <div class="feature">
        <h3>1. Browse Service Repositories</h3>
        <p>Navigate your orgs and repositories to select the services you want in the platform.</p>
      </div>
      
      <div class="feature">
        <h3>2. Submit PRs to Add Service</h3>
        <p>Trigger a PR from within the orchestrator UI to add that service to the platform configuration.</p>
      </div>
      
      <div class="feature">
        <h3>3. Merge PR to Trigger Workflow</h3>
        <p>Review and merge the PR on Github to see a real-time update of the platform services list.</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { onMounted, watch } from 'vue';

const router = useRouter();
const authStore = useAuthStore();
const title = import.meta.env.VITE_APP_TITLE
onMounted(() => {
  // Skip auth check if we've recently logged out or auth has failed
  if (authStore.recentlyLoggedOut || authStore.authenticationFailed) {
    return;
  }

  // Check auth state without triggering a new API call
  if (authStore.isAuthenticated) {
    router.push('/dashboard');
  }
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
  max-width: 1200px;
  margin: 0 auto;
  padding: 4rem 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.hero {
  text-align: center;
  margin-bottom: 4rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  max-width: 800px;
}

h1 {
  font-size: 2.5rem;
  margin-bottom: 1rem;
  color: var(--color-text-white);
}

.gradient-subtitle {
  font-size: 2rem;
  line-height: 1.3;
  font-weight: 500;
  margin-bottom: 2rem;
  color: var(--color-text-white);
}

p {
  color: var(--color-text-light);
  margin-bottom: 1.5rem;
  font-size: 1.1rem;
  line-height: 1.6;
}

.login-instructions {
  margin: 1.5rem 0;
  max-width: 600px;
}

.welcome-back {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding: 2rem;
  background-color: var(--color-bg-medium);
  border-radius: 8px;
  border: 1px solid var(--color-border);
  width: 100%;
  max-width: 500px;
}

.cta-button {
  display: inline-block;
  background: linear-gradient(90deg, var(--color-accent-blue), var(--color-primary));
  color: white;
  padding: 12px 24px;
  border-radius: 4px;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.3s;
  font-size: 1rem;
}

.cta-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  color: white;
}

.features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  width: 100%;
}

.feature {
  background-color: var(--color-bg-medium);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 2rem;
  transition: transform 0.3s, box-shadow 0.3s;
}

.feature:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 20px rgba(0, 0, 0, 0.1);
}

.feature h3 {
  font-size: 1.5rem;
  margin-bottom: 1rem;
  color: var(--color-primary);
}

.feature p {
  font-size: 1rem;
  color: var(--color-text-light);
  line-height: 1.5;
}

@media (max-width: 768px) {
  .home {
    padding: 2rem 1rem;
  }
  
  h1 {
    font-size: 2rem;
  }
  
  .gradient-subtitle {
    font-size: 1.5rem;
  }
}
</style>