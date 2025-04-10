<template>
  <div class="pull-requests">
    <div class="header-container">
      <h3>Pull Requests</h3>
      
      <div class="filter-container">
        <button 
          class="filter-pill" 
          :class="{ active: showOpen, open: showOpen }"
          @click="toggleFilter('open')"
        >
          Open
        </button>
        
        <button 
          class="filter-pill" 
          :class="{ active: showClosed, closed: showClosed }"
          @click="toggleFilter('closed')"
        >
          Closed
        </button>
      </div>
    </div>
    
    <div v-if="loading" class="loading-container">
      <div class="spinner"></div>
      <p>Loading pull requests...</p>
    </div>
    
    <div v-else-if="error" class="error-message">
      {{ error }}
    </div>
    
    <div v-else-if="pullRequests.length === 0" class="empty-message">
      No pull requests found
    </div>
    
    <div v-else-if="filteredPullRequests.length === 0" class="empty-message">
      No pull requests match the selected filters
    </div>
    
    <div v-else class="pr-list">
      <div 
        v-for="pr in filteredPullRequests" 
        :key="pr.id"
        class="pr-item"
      >
        <div class="pr-header">
          <span class="pr-number">#{{ pr.number }}</span>
          <span :class="['pr-status', pr.state]">{{ pr.state }}</span>
        </div>
        
        <h4 class="pr-title">{{ pr.title }}</h4>
        
        <div class="pr-meta">
          <div class="pr-author">
            <img 
              v-if="pr.user && pr.user.avatar_url" 
              :src="pr.user.avatar_url" 
              :alt="pr.user.login" 
              class="author-avatar"
            />
            <span>{{ pr.user ? pr.user.login : 'Unknown' }}</span>
          </div>
          <div class="pr-date">
            <span>Created: {{ formatDate(pr.created_at) }}</span>
          </div>
        </div>
        
        <a 
          :href="pr.html_url" 
          target="_blank" 
          rel="noopener noreferrer"
          class="pr-link"
        >
          View on GitHub
        </a>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useGithubStore } from '@/stores/github';
import api from '@/services/api';
import { useApiRequest } from '@/composables/useApiRequest';

const githubStore = useGithubStore();

// Local state for pull requests
const dataLoaded = ref(false); // Track if data has been loaded

// Use API request composable
const { loading, error, execute } = useApiRequest();

// Filter state
const showOpen = ref(true);
const showClosed = ref(false);

// Get pull requests from the store instead of local state
const pullRequests = computed(() => githubStore.pullRequests);

// Computed filtered pull requests
const filteredPullRequests = computed(() => {
  // If neither filter is selected, don't show any PRs
  if (!showOpen.value && !showClosed.value) {
    return [];
  }
  
  // Filter PRs based on selected states
  return pullRequests.value.filter(pr => {
    if (pr.state === 'open' && showOpen.value) return true;
    if (pr.state === 'closed' && showClosed.value) return true;
    return false;
  });
});

// Toggle filter method
const toggleFilter = (type) => {
  if (type === 'open') {
    showOpen.value = !showOpen.value;
  } else if (type === 'closed') {
    showClosed.value = !showClosed.value;
  }
};

// Format date helper
const formatDate = (dateString) => {
  if (!dateString) return '';
  const date = new Date(dateString);
  return date.toLocaleDateString();
};

// Load pull requests only once
const loadPullRequests = async () => {
  // Only load if not already loaded
  if (dataLoaded.value) return;
  
  // Use the store's method which handles caching
  await githubStore.fetchPullRequests();
  dataLoaded.value = true; // Mark as loaded
};

// Fetch pull requests when component mounts if not already loaded
onMounted(() => {
  loadPullRequests();
});
</script>
  
<style scoped>
.pull-requests {
  width: 100%;
  max-width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
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

.filter-container {
  display: flex;
  gap: 0.75rem;
}

.filter-pill {
  font-size: 12px;
  font-weight: 600;
  border-radius: 12px;
  padding: 2px 8px;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
  background-color: #eaecef;
  color: #6a737d;
  outline: none;
}

.filter-pill:hover {
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.filter-pill.open.active {
  background-color: rgba(46, 164, 79, 0.2);
  color: #2ea44f;
}

.filter-pill.closed.active {
  background-color: rgba(203, 36, 49, 0.2);
  color: #cb2431;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem 0;
  flex-grow: 1;
}

.spinner {
  width: 24px;
  height: 24px;
  border: 2px solid rgba(0, 0, 0, 0.1);
  border-radius: 50%;
  border-top-color: #0366d6;
  animation: spin 1s linear infinite;
  margin-bottom: 0.5rem;
}

.error-message {
  color: #cb2431;
  margin-top: 0.5rem;
  padding: 1rem;
  background-color: #ffdce0;
  border-radius: 6px;
  font-size: 14px;
  flex-grow: 1;
}

.empty-message {
  color: #586069;
  margin-top: 0.5rem;
  font-style: italic;
  font-size: 14px;
  padding: 2rem 0;
  text-align: center;
  background-color: #f6f8fa;
  border-radius: 6px;
  border: 1px dashed #e1e4e8;
  flex-grow: 1;
}

.pr-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  flex-grow: 1;
  overflow-y: auto;
  min-height: 0;
  padding-top: 5px;    /* Add padding to top */
  padding-bottom: 5px; /* Add padding to bottom */
  margin: -5px 0;      /* Negative margin to compensate for padding */
}

.pr-item {
  background-color: #f6f8fa;
  border: 1px solid #e1e4e8;
  border-radius: 6px;
  padding: 1rem;
  transition: transform 0.2s, box-shadow 0.2s;
}

.pr-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.pr-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.pr-number {
  color: #586069;
  font-size: 14px;
  font-weight: 600;
}

.pr-status {
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  text-transform: capitalize;
}

.pr-status.open {
  background-color: #2ea44f;
  color: white;
}

.pr-status.closed {
  background-color: #cb2431;
  color: white;
}

.pr-title {
  margin: 0.5rem 0;
  font-size: 16px;
  color: #24292e;
}

.pr-meta {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin: 0.75rem 0;
  font-size: 13px;
  color: #586069;
}

.pr-author {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.author-avatar {
  width: 20px;
  height: 20px;
  border-radius: 50%;
}

.pr-link {
  display: inline-block;
  margin-top: 0.5rem;
  color: #0366d6;
  text-decoration: none;
  font-size: 14px;
}

.pr-link:hover {
  text-decoration: underline;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (prefers-color-scheme: dark) {
  h3 {
    color: var(--color-text-white);
  }
  
  .pr-item {
    background-color: var(--color-bg-light);
    border-color: var(--color-border);
  }
  
  .pr-title {
    color: var(--color-text-white);
  }
  
  .pr-number, .pr-meta {
    color: var(--color-text-light);
  }
  
  .empty-message {
    background-color: var(--color-bg-light);
    border-color: var(--color-border);
    color: var(--color-text-light);
  }
}
</style>