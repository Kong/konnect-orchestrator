<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1 class="gradient-text">GitHub Repository Dashboard</h1>
    </div>

    <div class="dashboard-content">
      <div class="main-content">
        <div class="content-header">
          <div v-if="githubStore.currentOrg" class="current-context">
            <span v-if="githubStore.isPersonalAccount" class="context-badge personal">
              User
            </span>
            <span v-else class="context-badge org">
              Organization
            </span>
            <h2>{{ githubStore.currentOrg.name || githubStore.currentOrg.login }}</h2>
          </div>
        </div>

        <div class="repo-selector">
          <RepoDropdown ref="repoDropdown" @change="handleRepoChange" />

          <!-- Configuration Form placed outside repo details but before button -->
          <div v-if="githubStore.currentRepo" class="service-config-form">
            <h4>Service Configuration</h4>

            <!-- Prod Branch -->
            <div class="form-group">
              <label>Prod Branch:</label>
              <div class="editable-field">
                <input type="text" v-model="prodBranch" :readonly="!editingProdBranch" class="branch-input" />
                <button @click="toggleEditProdBranch" class="edit-button" :class="{ 'active': editingProdBranch }">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="14" height="14" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path>
                  </svg>
                </button>
              </div>
            </div>

            <!-- Dev Branch -->
            <div class="form-group">
              <label>Dev Branch:</label>
              <div class="editable-field">
                <input type="text" v-model="devBranch" :readonly="!editingDevBranch" class="branch-input" />
                <button @click="toggleEditDevBranch" class="edit-button" :class="{ 'active': editingDevBranch }">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="14" height="14" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path>
                  </svg>
                </button>
              </div>
            </div>

            <!-- Team Selection -->
            <div class="form-group">
              <label>Team:</label>
              <div v-if="!addingNewTeam">
                <select v-model="selectedTeam" class="team-input">
                  <option value="">-- Select Team --</option>
                  <option v-for="team in githubStore.availableTeams" :key="team" :value="team">
                    {{ team }}
                  </option>
                  <option value="new">+ Add New Team</option>
                </select>
              </div>
              <div v-else class="new-team-field">
                <input type="text" v-model="newTeamName" placeholder="Enter new team name" class="branch-input" />
                <div class="button-group">
                  <button @click="confirmNewTeam" class="confirm-button">
                    ✓
                  </button>
                  <button @click="cancelNewTeam" class="cancel-button">
                    ✕
                  </button>
                </div>
              </div>
            </div>
          </div>

          <button 
  v-if="githubStore.currentRepo" 
  @click="sendServiceRepoInfo" 
  class="add-service-btn"
  :disabled="sendingRepo || !isFormValid"
>
  <span v-if="sendingRepo" class="spinner-sm"></span>
  {{ sendingRepo ? 'Sending...' : 'Add Service to Platform' }}
</button>
<div v-if="sendError" class="error-message">
  {{ sendError }}
</div>
        </div>
      </div>

      <!-- Services Column -->
      <div class="services-column">
        <ServicesPanel ref="servicesPanel" />
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
import { onMounted, watch, ref, computed } from 'vue';
import api from '@/services/api';
import { useApiRequest } from '@/composables/useApiRequest';

const router = useRouter();
const authStore = useAuthStore();
const githubStore = useGithubStore();
const { loading: sendingRepo, error: sendError, execute } = useApiRequest();

const repoDropdown = ref(null);
const servicesPanel = ref(null);

// Service configuration state
const prodBranch = ref('prod');
const devBranch = ref('dev');
const selectedTeam = ref('');
const newTeamName = ref('');
const editingProdBranch = ref(false);
const editingDevBranch = ref(false);
const addingNewTeam = ref(false);

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

const toggleEditProdBranch = () => {
  editingProdBranch.value = !editingProdBranch.value;
};

const toggleEditDevBranch = () => {
  editingDevBranch.value = !editingDevBranch.value;
};

const sendServiceRepoInfo = async () => {
  if (!githubStore.currentRepo || !isFormValid.value) return;

  // Get the current repository object directly from the store
  const repo = githubStore.currentRepo;

  // Add is_enterprise field if it doesn't exist
  if (repo.is_enterprise === undefined) {
    repo.is_enterprise = false; // Default to false if not specified
  }

  // Execute API request using our composable
  const response = await execute(
    () => api.services.sendServiceRepoInfo(
      repo,
      addingNewTeam.value ? newTeamName.value : selectedTeam.value,
      prodBranch.value,
      devBranch.value
    ),
    'Failed to add service',
    null
  );

  if (response) {
    // Show success
    console.log('Repository registered successfully:', response);
    alert('Service added successfully!');

    // Refresh services list
    servicesPanel.value.fetchServices();
    
    // Invalidate pull requests cache since new PRs may have been created
    githubStore.invalidateCache('pullRequests');
    
    // Optionally, refresh pull requests in the background
    // This preloads the data for when the user views PRs next
    githubStore.fetchPullRequests(true).catch(err => {
      console.error('Background PR refresh failed:', err);
      // We don't show this error to the user since it's a background operation
    });
  }
};

const confirmNewTeam = () => {
  if (newTeamName.value) {
    // Add to available teams
    githubStore.setAvailableTeams([...githubStore.availableTeams, newTeamName.value]);
    // Select the new team
    selectedTeam.value = newTeamName.value;
    // Reset state
    addingNewTeam.value = false;
    newTeamName.value = '';
  }
};

const cancelNewTeam = () => {
  addingNewTeam.value = false;
  newTeamName.value = '';
};

// Watch for selected team changes
watch(selectedTeam, (newVal) => {
  if (newVal === 'new') {
    addingNewTeam.value = true;
    selectedTeam.value = '';
  }
});

const isFormValid = computed(() => {
  return prodBranch.value && devBranch.value && selectedTeam.value;
});

</script>

<style scoped>
.dashboard {
  width: 100%;
  max-width: 1600px;
}

.dashboard-header {
  margin-bottom: 2.5rem;
  position: relative;
}

.dashboard-header::after {
  content: "";
  position: absolute;
  bottom: -10px;
  left: 0;
  width: 100px;
  height: 4px;
  background: linear-gradient(90deg, var(--color-accent-blue), var(--color-primary));
  border-radius: 2px;
}

h1 {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 1rem;
}

.dashboard-content {
  display: grid;
  grid-template-columns: 450px 1fr 300px;
  /* Wider left column, narrower middle column */
  gap: 20px;
  /* Consistent gap between all columns */
  width: 100%;
  margin: 0 auto;
  background-color: var(--color-bg-dark);
  align-items: start;
  /* Align columns to the top */
}

.main-content,
.pull-requests-column {
  width: 100%;
  background-color: var(--color-bg-medium);
  border-radius: 8px;
  padding: 1.5rem;
  border: 1px solid var(--color-border);
  box-sizing: border-box;
  /* Ensure padding is included in the width */
  height: auto;
  /* Auto height based on content */
  min-height: 200px;
  /* Minimum height to ensure visibility */
}

.services-column {
  width: 100%;
  background-color: var(--color-bg-medium);
  border-radius: 8px;
  padding: 1.5rem;
  border: 1px solid var(--color-border);
  box-sizing: border-box;
  position: relative; /* Create stacking context */
  overflow: visible;  /* Allow children to overflow */
}

/* If you have a .services-panel deep selector, update it too */
.services-column :deep(.services-panel) {
  padding-right: 10px;
  overflow: visible; /* Allow overflow */
  position: relative; /* Create stacking context */
}

.repo-selector {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.add-service-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 12px;
  background: linear-gradient(90deg, var(--color-accent-blue), var(--color-primary));
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.add-service-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.add-service-btn:disabled {
  background: linear-gradient(90deg, #666, #999);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
  opacity: 0.7;
}

.spinner-sm {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s linear infinite;
}

.service-config-form {
  margin: 1rem 0;
  padding: 1rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background-color: var(--color-bg-light);
}

.service-config-form h4 {
  margin: 0 0 1rem 0;
  font-size: 14px;
  color: var(--color-text-white);
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.25rem;
  font-size: 13px;
  color: var(--color-text-light);
}

.form-group:last-child {
  margin-bottom: 0;
}

.editable-field {
  display: flex;
  align-items: center;
}

.branch-input,
.team-input {
  flex: 1;
  padding: 8px 10px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background-color: var(--color-bg-dark);
  color: var(--color-text-white);
  font-size: 13px;
}

.branch-input:read-only {
  opacity: 0.8;
}

.edit-button {
  background: none;
  border: none;
  color: var(--color-text-muted);
  padding: 5px;
  margin-left: 5px;
  cursor: pointer;
  border-radius: 4px;
  line-height: 0;
}

.edit-button:hover,
.edit-button.active {
  color: var(--color-primary);
  background-color: rgba(0, 185, 105, 0.1);
}

.new-team-field {
  display: flex;
  align-items: center;
}

.button-group {
  display: flex;
  margin-left: 8px;
}

.confirm-button,
.cancel-button {
  background: none;
  border: none;
  color: var(--color-text-white);
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-radius: 4px;
  margin-left: 4px;
}

.confirm-button {
  background-color: var(--color-primary);
}

.confirm-button:hover {
  background-color: #00a057;
}

.cancel-button {
  background-color: var(--color-bg-light);
}

.cancel-button:hover {
  background-color: #f85149;
  color: white;
}

.content-header {
  margin-bottom: 1.5rem;
}

.current-context {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.context-badge {
  padding: 6px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.context-badge.personal {
  background-color: var(--color-accent-blue);
  color: white;
}

.context-badge.org {
  background-color: var(--color-primary);
  color: white;
}

.current-context h2 {
  margin: 0;
  font-size: 1.4rem;
  color: var(--color-text-white);
}

@media (max-width: 1200px) {
  .dashboard-content {
    width: 100%;
    gap: 20px;
    /* Keep consistent gap even on smaller screens */
  }
}

@media (max-width: 992px) {
  .dashboard {
    width: 100%;
  }

  .dashboard-content {
    grid-template-columns: 1fr;
    gap: 20px;
    /* Maintain vertical gap between stacked columns */
  }

  .services-column,
  .pull-requests-column {
    position: static;
    max-height: none;
    overflow-y: visible;
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>