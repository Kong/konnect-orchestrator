<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1 class="gradient-text">{{ title }}</h1>
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

           <!-- Production Branch -->
<div class="form-group">
  <label for="prod-branch">Production Branch</label>
  <select 
    id="prod-branch" 
    v-model="prodBranch"
    class="form-control"
    :disabled="!githubStore.branches.length || sendingRepo"
  >
    <option value="" disabled>Select production branch</option>
    <option 
      v-for="branch in githubStore.branchOptions" 
      :key="branch.value" 
      :value="branch.value"
      :class="{'default-branch': branch.isDefault}"
    >
      {{ branch.label }}
    </option>
  </select>
</div>

<!-- Development Branch -->
<div class="form-group">
  <label for="dev-branch">Development Branch</label>
  <select 
    id="dev-branch" 
    v-model="devBranch"
    class="form-control"
    :disabled="!githubStore.branches.length || sendingRepo"
  >
    <option value="" disabled>Select development branch</option>
    <option 
      v-for="branch in githubStore.branchOptions" 
      :key="branch.value" 
      :value="branch.value"
      :class="{'default-branch': branch.isDefault}"
    >
      {{ branch.label }}
    </option>
  </select>
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

          <button v-if="githubStore.currentRepo" @click="sendServiceRepoInfo" class="add-service-btn"
            :disabled="sendingRepo || !isFormValid">
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
import { useApiRequest } from '@/composables/useApiRequest';
import { useToast } from '@/composables/useToast';


const router = useRouter();
const authStore = useAuthStore();
const githubStore = useGithubStore();
const { loading: sendingRepo, error: sendError, execute } = useApiRequest();
const title = import.meta.env.VITE_APP_TITLE
const repoDropdown = ref(null);
const servicesPanel = ref(null);

// Service configuration state
const prodBranch = ref('prod');
const devBranch = ref('dev');
const selectedTeam = ref('');
const newTeamName = ref('');
const addingNewTeam = ref(false);

const toast = useToast();

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

watch(sendError, (error) => {
  if (error) {
    toast.error(`Failed to add service: ${error}`);
  }
});

// Methods
const handleRepoChange = (repoFullName) => {
  console.log('Selected repository:', repoFullName);
};

const sendServiceRepoInfo = async () => {
  if (!githubStore.currentRepo || !isFormValid.value) return;

  // Get the current repository object directly from the store
  const repo = githubStore.currentRepo;
  
  // Determine which team to use
  const teamName = addingNewTeam.value ? newTeamName.value : selectedTeam.value;

  try {
    // Execute using the store action instead of directly calling the API
    await execute(
      () => githubStore.registerService(
        repo, 
        teamName,
        prodBranch.value,
        devBranch.value
      ),
      'Failed to add service',
      null
    );
    
    // Show success
    toast.success('Service added successfully!');

    // Refresh services list
    servicesPanel.value.fetchServices();
    
    // Optionally, refresh pull requests in the background
    githubStore.fetchPullRequests(true).catch(err => {
      console.error('Background PR refresh failed:', err);
      toast.warning('Background pull request refresh failed. Try refreshing the page if updates are missing.');
    });
  } catch (error) {
    // Error handling happens in the execute function
    console.error('Service registration error:', error);
    // The execute composable will show errors automatically
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

watch(() => githubStore.selectedRepo, async (newRepo) => {
  if (newRepo) {
    try {
      // Branches will be fetched automatically by the store
      // Wait for branches to be loaded
      if (githubStore.loading.value) {
        await new Promise(resolve => setTimeout(resolve, 500));
      }
      
      // Set default values based on repository branches
      if (githubStore.branches.length > 0) {
        // Find default branch
        const defaultBranch = githubStore.branches.find(branch => branch.is_default);
        if (defaultBranch) {
          prodBranch.value = defaultBranch.name;
          
          // Look for common dev branch names
          const devBranchNames = ['dev', 'develop', 'development', 'staging', 'test'];
          const devBranch = githubStore.branches.find(branch => 
            devBranchNames.includes(branch.name.toLowerCase())
          );
          
          if (devBranch) {
            devBranch.value = devBranch.name;
          } else if (prodBranch.value !== defaultBranch.name) {
            // Use default branch as dev if it's not already used for prod
            devBranch.value = defaultBranch.name;
          } else {
            // If we can't find a good dev branch, just use the same as prod
            devBranch.value = prodBranch.value;
          }
        }
      }
    } catch (error) {
      toast.error('Failed to load repository branches');
      console.error('Error fetching branches:', error);
    }
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

.default-branch {
  font-weight: 600;
  color: #646cff;
}

select.form-control {
  width: 100%;
  padding: 0.6em 1.2em;
  font-size: 1em;
  font-weight: 500;
  font-family: inherit;
  background-color: #1a1a1a;
  border-radius: 8px;
  border: 1px solid transparent;
  transition: border-color 0.25s;
}

select.form-control:hover {
  border-color: #646cff;
}

select.form-control:focus,
select.form-control:focus-visible {
  outline: 4px auto -webkit-focus-ring-color;
}

@media (prefers-color-scheme: light) {
  select.form-control {
    background-color: #f9f9f9;
  }
}

.services-column {
  width: 100%;
  background-color: var(--color-bg-medium);
  border-radius: 8px;
  padding: 1.5rem;
  border: 1px solid var(--color-border);
  box-sizing: border-box;
  position: relative;
  /* Create stacking context */
  overflow: visible;
  /* Allow children to overflow */
}

/* If you have a .services-panel deep selector, update it too */
.services-column :deep(.services-panel) {
  padding-right: 10px;
  overflow: visible;
  /* Allow overflow */
  position: relative;
  /* Create stacking context */
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