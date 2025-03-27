<template>
  <div class="repo-dropdown">
    <label for="repo-select">Select Repository:</label>
    
    <div class="select-container">
      <select 
        id="repo-select"
        v-model="selectedValue"
        :disabled="githubStore.loading || githubStore.repositories.length === 0"
        @change="handleChange"
      >
        <option value="">-- Select a repository --</option>
        <option 
          v-for="option in githubStore.repoOptions" 
          :key="option.value" 
          :value="option.value"
        >
          {{ option.label }} {{ option.isPrivate ? '🔒' : '' }}
        </option>
      </select>
      
      <div v-if="githubStore.loading" class="spinner"></div>
    </div>
    
    <div v-if="githubStore.error" class="error-message">
      {{ githubStore.error }}
    </div>
    
    <div v-if="!githubStore.loading && githubStore.repositories.length === 0" class="empty-message">
      No repositories found in this organization
    </div>
    
    <!-- Display selected repository details -->
    <div v-if="githubStore.currentRepo" class="repo-details">
      <h3>{{ githubStore.currentRepo.name }}</h3>
      <p v-if="githubStore.currentRepo.description">{{ githubStore.currentRepo.description }}</p>
      <div class="repo-meta">
        <span>{{ githubStore.currentRepo.private ? 'Private' : 'Public' }}</span>
        <span>Owner: {{ githubStore.currentRepo.owner.login }}</span>
      </div>
      
      <a 
        :href="githubStore.currentRepo.html_url" 
        target="_blank" 
        rel="noopener noreferrer"
        class="repo-link"
      >
        View on GitHub
      </a>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';
import { useGithubStore } from '@/stores/github';
import { useAuthStore } from '@/stores/auth';

// Emit events
const emit = defineEmits(['change']);

// Store
const githubStore = useGithubStore();
const authStore = useAuthStore();

// State
const selectedValue = ref('');

// Methods
const handleChange = () => {
  githubStore.selectRepository(selectedValue.value);
  emit('change', selectedValue.value);
};

// Watch for authentication changes
watch(() => authStore.isAuthenticated, (isAuthenticated) => {
  if (!isAuthenticated) {
    githubStore.reset();
    selectedValue.value = '';
  }
});

// Watch for repository list changes
watch(() => githubStore.repositories, (newRepos) => {
  // If we have a selected repository in the store
  if (githubStore.selectedRepo) {
    // Check if it exists in the new repository list
    if (newRepos.some(repo => repo.full_name === githubStore.selectedRepo)) {
      // Update the dropdown to match
      if (selectedValue.value !== githubStore.selectedRepo) {
        selectedValue.value = githubStore.selectedRepo;
      }
    } else {
      // If the repository is no longer available, clear the selection
      selectedValue.value = '';
    }
  } else {
    // If no repository is selected in the store, clear the dropdown
    selectedValue.value = '';
  }
});

// Watch for changes to the selectedRepo in the store
watch(() => githubStore.selectedRepo, (newRepo) => {
  // Only update if different to avoid loops
  if (newRepo !== selectedValue.value) {
    selectedValue.value = newRepo || '';
  }
});

// Synchronize with the store when mounting
onMounted(() => {
  if (githubStore.selectedRepo) {
    selectedValue.value = githubStore.selectedRepo;
  }
});
</script>

<style scoped>
.repo-dropdown {
  margin: 1rem 0;
  width: 100%;
}

label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: var(--color-text-light);
}

.select-container {
  position: relative;
  display: flex;
  width: 100%;
}

select {
  width: 100%;
  padding: 10px 36px 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background-color: var(--color-bg-light);
  color: var(--color-text-white);
  font-size: 14px;
  appearance: none;
  background-image: url("data:image/svg+xml;charset=utf-8,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='none' viewBox='0 0 12 12'%3E%3Cpath fill='%23fff' d='M6 8.825c-.2 0-.4-.1-.5-.2l-3.8-3.6c-.3-.3-.3-.8 0-1.1.3-.3.7-.3 1 0L6 7.025l3.3-3.1c.3-.3.7-.3 1 0 .3.3.3.8 0 1.1l-3.8 3.6c-.1.1-.3.2-.5.2Z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  cursor: pointer;
}

select:disabled {
  background-color: var(--color-bg-dark);
  opacity: 0.6;
  cursor: not-allowed;
}

.spinner {
  position: absolute;
  right: 40px;
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: var(--color-primary);
  animation: spin 1s linear infinite;
}

.error-message {
  color: #f85149;
  margin-top: 0.5rem;
  font-size: 14px;
}

.empty-message {
  color: var(--color-text-muted);
  margin-top: 0.5rem;
  font-style: italic;
  font-size: 14px;
}

.repo-details {
  margin-top: 1.5rem;
  padding: 1rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background-color: var(--color-bg-light);
}

.repo-details h3 {
  margin: 0 0 0.5rem 0;
  font-size: 16px;
  color: var(--color-text-white);
}

.repo-details p {
  margin: 0 0 1rem 0;
  color: var(--color-text-light);
  font-size: 14px;
}

.repo-meta {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
  font-size: 13px;
  color: var(--color-text-light);
}

.repo-link {
  display: inline-block;
  color: var(--color-accent-blue);
  text-decoration: none;
  font-size: 14px;
}

.repo-link:hover {
  color: var(--color-primary);
  text-decoration: underline;
}

@keyframes spin {
  to { transform: translateY(-50%) rotate(360deg); }
}
</style>