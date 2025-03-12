<template>
    <div class="dashboard">
      <h1>GitHub Repository Dashboard</h1>
      
      <div class="dashboard-content">
        <div class="sidebar">
          <UserProfile />
          
          <div class="filter-options">
            <h3>Organizations & Repositories</h3>
            <OrgDropdown @change="handleOrgChange" />
          </div>
        </div>
        
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
          
          <RepoDropdown @change="handleRepoChange" />
          
          <div v-if="githubStore.currentRepo" class="repo-content">
            <div class="repo-stats">
              <div class="stat-item">
                <h4>Visibility</h4>
                <div>{{ githubStore.currentRepo.private ? 'Private' : 'Public' }}</div>
              </div>
              <div class="stat-item">
                <h4>Default Branch</h4>
                <div>{{ githubStore.currentRepo.default_branch }}</div>
              </div>
              <div class="stat-item">
                <h4>Created</h4>
                <div>{{ formatDate(githubStore.currentRepo.created_at) }}</div>
              </div>
              <div class="stat-item">
                <h4>Updated</h4>
                <div>{{ formatDate(githubStore.currentRepo.updated_at) }}</div>
              </div>
            </div>
            
            <!-- Placeholder for repository content -->
            <div class="repo-placeholder">
              <p>Repository selected: {{ githubStore.currentRepo.name }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import { ref, computed } from 'vue';
  import { useRouter } from 'vue-router';
  import { useAuthStore } from '@/stores/auth';
  import { useGithubStore } from '@/stores/github';
  import UserProfile from '@/components/UserProfile.vue';
  import RepoDropdown from '@/components/RepoDropdown.vue';
  import OrgDropdown from '@/components/OrgDropdown.vue';
  import { onMounted, watch } from 'vue';
  
  const router = useRouter();
  const authStore = useAuthStore();
  const githubStore = useGithubStore();
  
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
  const handleOrgChange = (orgLogin) => {
    console.log('Selected organization:', orgLogin);
    // No need to do anything else as the store handles updating repositories
  };
  
  const handleRepoChange = (repoFullName) => {
    console.log('Selected repository:', repoFullName);
    // Additional actions when repo changes can be added here
  };
  
  const formatDate = (dateString) => {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString();
  };
  </script>
  
  <style scoped>
  .dashboard {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
  }
  
  h1 {
    margin-bottom: 2rem;
    color: #24292e;
  }
  
  .dashboard-content {
    display: grid;
    grid-template-columns: 300px 1fr;
    gap: 2rem;
  }
  
  .sidebar {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }
  
  .filter-options {
    padding: 1.5rem;
    background-color: #f6f8fa;
    border: 1px solid #e1e4e8;
    border-radius: 6px;
  }
  
  .filter-options h3 {
    margin-top: 0;
    margin-bottom: 1rem;
    font-size: 16px;
  }
  
  .main-content {
    width: 100%;
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
  
  .repo-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
  }
  
  .stat-item {
    padding: 1rem;
    background-color: #f6f8fa;
    border: 1px solid #e1e4e8;
    border-radius: 6px;
  }
  
  .stat-item h4 {
    margin: 0 0 0.5rem 0;
    font-size: 14px;
    color: #586069;
  }
  
  .stat-item div {
    font-size: 16px;
    font-weight: 600;
    color: #24292e;
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
  
  @media (max-width: 768px) {
    .dashboard-content {
      grid-template-columns: 1fr;
    }
  }
  </style>