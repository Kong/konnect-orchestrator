<template>
  <div class="dashboard">
    <h1>GitHub Repository Dashboard</h1>
    
    <div class="dashboard-content">
      <div class="main-content">
        <div class="content-header">
          <div v-if="githubStore.currentOrg" class="current-context">
            <span v-if="githubStore.isPersonalAccount" class="context-badge personal">
              Personal Account
            </span>
            <span v-else class="context-badge org">
              Organization
            </span>
            <h2>{{ githubStore.currentOrg.name || githubStore.currentOrg.login }}</h2>
          </div>
        </div>
        
        <div class="repo-selector">
          <RepoDropdown @change="handleRepoChange" />
          
          <button 
            v-if="githubStore.currentRepo" 
            @click="sendServiceRepoInfo" 
            class="send-repo-btn"
            :disabled="sendingRepo"
          >
            <span v-if="sendingRepo" class="spinner-sm"></span>
            {{ sendingRepo ? 'Sending...' : 'Send Repository Info' }}
          </button>
        </div>
      </div>
      
      <!-- Services Column -->
      <div class="services-column">
        <ServicesPanel />
      </div>
      
      <!-- Pull Requests Column -->
      <div class="pull-requests-column">
        <PullRequestList />
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useGithubStore } from '@/stores/github';
import RepoDropdown from '@/components/RepoDropdown.vue';
import PullRequestList from '@/components/PullRequestList.vue';
import ServicesPanel from '@/components/ServicesPanel.vue';
import { onMounted, watch, ref } from 'vue';
import api from '@/services/api';

const router = useRouter();
const authStore = useAuthStore();
const githubStore = useGithubStore();
const sendingRepo = ref(false);

// Check authentication on component mount
onMounted(() => {
  if (!authStore.isAuthenticated) {
    router.push('/');
  } else {
    authStore.loadUser();
    githubStore.fetchOrganizations();
  }
});

// Watch for authentication changes
watch(() => authStore.isAuthenticated, (isAuthenticated) => {
  if (!isAuthenticated) {
    router.push('/');
  }
});

// Methods
const handleRepoChange = (repoFullName) => {
  console.log('Selected repository:', repoFullName);
};

const sendServiceRepoInfo = async () => {
  if (!githubStore.currentRepo) return;
  
  sendingRepo.value = true;
  
  try {
    // Get the current repository object directly from the store
    const repo = githubStore.currentRepo;
    
    // Add is_enterprise field if it doesn't exist
    if (repo.is_enterprise === undefined) {
      repo.is_enterprise = false; // Default to false if not specified
    }
    
    // Send to backend using updated API method that matches the Go struct
    const response = await api.services.sendServiceRepoInfo(repo, "core");
    
    // Show success (could use a toast notification here)
    console.log('Repository registered successfully:', response.data);
    
    alert('Repository registered successfully!');
  } catch (error) {
    console.error('Error registering repository:', error);
    alert('Failed to register repository. See console for details.');
  } finally {
    sendingRepo.value = false;
  }
};
</script>

<style scoped>
.dashboard {
  width: 100%;
  max-width: 1600px;
  margin: 0 auto;
  padding: 2rem;
}

h1 {
  margin-bottom: 2rem;
  color: #24292e;
}

.dashboard-content {
  display: grid;
  grid-template-columns: 300px 2fr 300px; /* Three-column layout with wider middle column */
  gap: 2rem;
  width: 85%; /* Set to 85% of the page width for better use of space */
  margin: 0 auto;
}

.main-content {
  width: 100%;
}

.repo-selector {
  position: sticky;
  top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.send-repo-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 10px;
  background-color: #2ea44f;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.send-repo-btn:hover {
  background-color: #2c974b;
}

.send-repo-btn:disabled {
  background-color: #94d3a2;
  cursor: not-allowed;
}

.spinner-sm {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s linear infinite;
}

.services-column {
  width: 100%;
  border-left: 1px solid #e1e4e8;
  padding-left: 1.5rem;
}

.pull-requests-column {
  width: 100%;
  border-left: 1px solid #e1e4e8;
  padding-left: 1.5rem;
  position: sticky;
  top: 1rem;
  align-self: start;
  max-height: calc(100vh - 4rem);
  overflow-y: auto;
}

.content-header {
  margin-bottom: 1.5rem;
}

.current-context {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.context-badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.context-badge.personal {
  background-color: #0366d6;
  color: white;
}

.context-badge.org {
  background-color: #2ea44f;
  color: white;
}

.current-context h2 {
  margin: 0;
  font-size: 1.4rem;
}

.repo-content {
  margin-top: 2rem;
}

.repo-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  background-color: #f6f8fa;
  border: 1px solid #e1e4e8;
  border-radius: 6px;
  color: #586069;
  font-style: italic;
}

@media (max-width: 1200px) {
  .dashboard-content {
    width: 95%;
  }
}

@media (max-width: 992px) {
  .dashboard {
    width: 100%;
    padding: 1rem;
  }
  
  .dashboard-content {
    width: 100%;
    grid-template-columns: 1fr;
  }
  
  .services-column,
  .pull-requests-column {
    border-left: none;
    padding-left: 0;
    border-top: 1px solid #e1e4e8;
    padding-top: 1.5rem;
    margin-top: 1.5rem;
    position: static;
    max-height: none;
    overflow-y: visible;
  }
}
</style>