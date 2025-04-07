<template>
  <div class="services-panel">
    <div class="header-container">
      <h3>Services</h3>
      
      <div class="team-filter">
        <select 
          v-model="selectedTeam"
          class="team-select"
        >
          <option value="">All Teams</option>
          <option 
            v-for="team in teams" 
            :key="team" 
            :value="team"
          >
            {{ team }}
          </option>
        </select>
      </div>
    </div>
    
    <div v-if="loading" class="loading-container">
      <div class="spinner"></div>
      <p>Loading services...</p>
    </div>
    
    <div v-else-if="error" class="error-message">
      {{ error }}
    </div>
    
    <div v-else-if="filteredServices.length === 0" class="empty-message">
      No services found
    </div>
    
    <div v-else class="services-list">
      <div 
        v-for="service in filteredServices" 
        :key="service.name"
        class="service-card"
      >
        <h4 class="service-name">{{ service.name }}</h4>
        <p v-if="service.description" class="service-description">{{ service.description }}</p>
        
        <div class="service-meta">
          <div class="service-team">
            <span class="team-label">Team:</span>
            <span class="team-name">{{ service.team }}</span>
          </div>
          <div class="service-branch">
            <span class="branch-label">Prod:</span>
            <span class="branch-name">{{ service.prodBranch }}</span>
          </div>
          <div class="service-branch">
            <span class="branch-label">Dev:</span>
            <span class="branch-name">{{ service.devBranch }}</span>
          </div>
          <div v-if="service.specPath" class="service-path">
            <span class="path-label">Spec:</span>
            <span class="path-value">{{ service.specPath }}</span>
          </div>
        </div>
        
        <div v-if="service.git && service.git.repo" class="service-repo">
          <a 
            :href="getRepoUrl(service.git)" 
            target="_blank" 
            rel="noopener noreferrer"
            class="repo-link"
          >
            View Repository
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue';
import api from '@/services/api';
import { useGithubStore } from '@/stores/github';
import { useApiRequest } from '@/composables/useApiRequest';

// State
// First define all state variables and composables
const services = ref([]);
const teams = ref([]);
const selectedTeam = ref('');
const githubStore = useGithubStore();
const { loading, error, execute } = useApiRequest();
const lastFetchTime = ref(null);
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

// Define computed properties
const filteredServices = computed(() => {
  if (!selectedTeam.value) {
    return services.value;
  }
  return services.value.filter(service => service.team === selectedTeam.value);
});

// Define functions
const getTeams = () => teams.value;

const fetchServices = async (forceRefresh = false) => {
  // Skip if data is fresh and not forcing refresh
  if (
    !forceRefresh && 
    services.value.length > 0 && 
    lastFetchTime.value && 
    (Date.now() - lastFetchTime.value < CACHE_DURATION)
  ) {
    console.log('Using cached services data');
    return services.value;
  }

  // Execute API request
  const response = await execute(
    () => api.services.getServices(),
    'Failed to load services',
    { services: [] }
  );
  
  // Update cache timestamp
  lastFetchTime.value = Date.now();
  
  services.value = response.services || [];
  
  // Extract unique teams
  const uniqueTeams = new Set(services.value.map(service => service.team).filter(Boolean));
  teams.value = Array.from(uniqueTeams).sort();
  
  // Notify the github store of available teams
  githubStore.setAvailableTeams(teams.value);
  
  return services.value;
};

const getRepoUrl = (git) => {
  if (!git || !git.repo) return '#';
  
  // Handle different git providers
  if (git.repo.startsWith('https://') || git.repo.startsWith('http://')) {
    return git.repo;
  }
  
  // Assume GitHub if no full URL provided
  return `https://github.com/${git.repo}`;
};

// Now expose functions AFTER they've been defined
defineExpose({ getTeams, fetchServices });

// Fetch services on component mount
onMounted(() => {
  fetchServices();
});
</script>

<style scoped>
.services-panel {
  width: 100%;
  max-width: 100%;
}

.header-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

h3 {
  margin: 0;
  font-size: 18px;
  color: var(--color-text-white);
}

.team-filter {
  min-width: 120px;
}

.team-select {
  background-color: var(--color-bg-light);
  color: var(--color-text-white);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 6px 24px 6px 10px;
  font-size: 12px;
  appearance: none;
  background-image: url("data:image/svg+xml;charset=utf-8,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='none' viewBox='0 0 12 12'%3E%3Cpath fill='%23fff' d='M6 8.825c-.2 0-.4-.1-.5-.2l-3.8-3.6c-.3-.3-.3-.8 0-1.1.3-.3.7-.3 1 0L6 7.025l3.3-3.1c.3-.3.7-.3 1 0 .3.3.3.8 0 1.1l-3.8 3.6c-.1.1-.3.2-.5.2Z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 8px center;
  cursor: pointer;
  transition: border-color 0.3s;
}

.team-select:hover, .team-select:focus {
  border-color: var(--color-primary);
  outline: none;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem 0;
}

.spinner {
  width: 24px;
  height: 24px;
  border: 2px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top-color: var(--color-primary);
  animation: spin 1s linear infinite;
  margin-bottom: 0.5rem;
}

.error-message {
  color: #f85149;
  margin-top: 0.5rem;
  padding: 1rem;
  background-color: rgba(248, 81, 73, 0.1);
  border-radius: 6px;
  font-size: 14px;
}

.empty-message {
  color: var(--color-text-muted);
  margin-top: 0.5rem;
  font-style: italic;
  font-size: 14px;
  padding: 2rem 0;
  text-align: center;
  background-color: var(--color-bg-light);
  border-radius: 6px;
  border: 1px dashed var(--color-border);
}

/* Update the .services-list class */
.services-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-height: calc(100vh - 12rem);
  overflow-y: auto;
  padding-right: 10px;
  padding-top: 5px;    /* Add padding to top */
  padding-bottom: 5px; /* Add padding to bottom */
  margin: -5px 0;      /* Negative margin to compensate for padding */
}

/* Update service-card styles */
.service-card {
  background-color: var(--color-bg-light);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 1rem;
  position: relative;
  z-index: 1;
  will-change: transform; /* Performance optimization for transforms */
  transition: transform 0.2s, box-shadow 0.2s, border-color 0.2s;
}

.service-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
  border-color: var(--color-primary);
  z-index: 5;
}

.service-name {
  margin: 0 0 0.75rem 0;
  font-size: 16px;
  color: var(--color-text-white);
}

.service-description {
  margin: 0 0 1rem 0;
  color: var(--color-text-light);
  font-size: 14px;
  line-height: 1.5;
}

.service-meta {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1rem;
  font-size: 13px;
}

.service-branch, .service-path, .service-team {
  display: flex;
  align-items: center;
}

.branch-label, .path-label, .team-label {
  font-weight: 600;
  color: var(--color-text-white);
  width: 45px;
}

.branch-name, .path-value, .team-name {
  color: var(--color-text-light);
}

.team-name {
  color: var(--color-primary);
  font-weight: 500;
}

.service-repo {
  margin-top: 0.75rem;
}

.repo-link {
  display: inline-block;
  color: var(--color-accent-blue);
  text-decoration: none;
  font-size: 14px;
  transition: color 0.2s;
}

.repo-link:hover {
  color: var(--color-primary);
  text-decoration: underline;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>