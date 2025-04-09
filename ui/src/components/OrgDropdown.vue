<template>
  <div class="org-dropdown">
    <label for="org-select">Select Account or Organization:</label>
    
    <div class="select-container">
      <select 
        id="org-select"
        v-model="selectedValue"
        :disabled="githubStore.loading || githubStore.organizations.length === 0"
        @change="handleChange"
      >
        <option value="">-- Select an account --</option>
        <option 
          v-for="option in githubStore.orgOptions" 
          :key="option.value" 
          :value="option.value"
          :class="{ 'personal-option': option.isPersonal }"
        >
          {{ option.label }}
        </option>
      </select>
      
      <div v-if="githubStore.loading" class="spinner"></div>
    </div>
    
    <div v-if="githubStore.error" class="error-message">
      {{ githubStore.error }}
    </div>
    
    <div v-if="!githubStore.loading && githubStore.organizations.length === 0" class="empty-message">
      No organizations found
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue';
import { useGithubStore } from '@/stores/github';
import { useAuthStore } from '@/stores/auth';

// Emit events
const emit = defineEmits(['change']);

// Store
const githubStore = useGithubStore();
const authStore = useAuthStore();

// State
const selectedValue = ref('');

const loadOrganizations = async (forceRefresh = false) => {
  if (authStore.isAuthenticated) {
    // Pass forceRefresh to use cache when appropriate
    await githubStore.fetchOrganizations(forceRefresh);
    
    // The selectedValue should reflect what's in the store
    if (githubStore.selectedOrg) {
      selectedValue.value = githubStore.selectedOrg;
    }
  }
};

const handleChange = () => {
  githubStore.selectOrganization(selectedValue.value);
  emit('change', selectedValue.value);
};

// Watch for authentication changes
watch(() => authStore.isAuthenticated, (isAuthenticated) => {
  if (isAuthenticated) {
    loadOrganizations();
  } else {
    githubStore.reset();
    selectedValue.value = '';
  }
}, { immediate: true });

// Watch for selectedOrg changes in store
watch(() => githubStore.selectedOrg, (newOrg) => {
  if (newOrg !== selectedValue.value) {
    selectedValue.value = newOrg || '';
  }
});

// Add a refresh method that can be called from parent
const refreshOrganizations = () => {
  return loadOrganizations(true); // Force refresh
};

defineExpose({ refreshOrganizations });

// Load organizations on component mount
onMounted(() => {
  if (authStore.isAuthenticated) {
    loadOrganizations();
  }
});
</script>

<style scoped>
.org-dropdown {
  margin: 1rem 0;
  width: 100%;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}

label {
  display: block;
  margin-bottom: 0.5rem;
  margin-right: 10px;
  font-weight: 500;
  color: var(--color-text-light);
  white-space: nowrap;
  width: 100%;
}

.select-container {
  position: relative;
  display: inline-flex;
  flex: 1;
  min-width: 0; /* Allow container to shrink */
}

select {
  width: 100%;
  padding: 10px 36px 10px 12px; /* Increased right padding to prevent text overlap with dropdown arrow */
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background-color: var(--color-bg-light);
  color: var(--color-text-white);
  font-size: 14px;
  appearance: none;
  background-image: url("data:image/svg+xml;charset=utf-8,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='none' viewBox='0 0 12 12'%3E%3Cpath fill='%23fff' d='M6 8.825c-.2 0-.4-.1-.5-.2l-3.8-3.6c-.3-.3-.3-.8 0-1.1.3-.3.7-.3 1 0L6 7.025l3.3-3.1c.3-.3.7-.3 1 0 .3.3.3.8 0 1.1l-3.8 3.6c-.1.1-.3.2-.5.2Z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  background-size: 12px;
  cursor: pointer;
  transition: border-color 0.3s;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

select:hover, select:focus {
  border-color: var(--color-primary);
  outline: none;
}

select:disabled {
  background-color: var(--color-bg-dark);
  opacity: 0.6;
  cursor: not-allowed;
}

.personal-option {
  font-weight: 600;
  color: var(--color-accent-blue);
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
  width: 100%;
}

.empty-message {
  color: var(--color-text-muted);
  margin-top: 0.5rem;
  font-style: italic;
  font-size: 14px;
  width: 100%;
}

@keyframes spin {
  to { transform: translateY(-50%) rotate(360deg); }
}

@media (max-width: 768px) {
  .org-dropdown {
    flex-direction: column;
    align-items: flex-start;
  }
  
  label {
    margin-bottom: 0.5rem;
    margin-right: 0;
  }
  
  .select-container {
    width: 100%;
  }
}
</style>