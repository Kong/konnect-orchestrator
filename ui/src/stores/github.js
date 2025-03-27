// src/stores/github.js
import { defineStore } from 'pinia';
import { ref, computed, watch } from 'vue';
import api from '@/services/api';
import { useAuthStore } from '@/stores/auth';
import DOMPurify from 'dompurify';
import { useApiRequest } from '@/composables/useApiRequest';

export const useGithubStore = defineStore('github', () => {
  // Get auth store for access to user info
  const authStore = useAuthStore();
  
  // Create API request composables for different endpoints
  const organizationsRequest = useApiRequest();
  const repositoriesRequest = useApiRequest();
  const pullRequestsRequest = useApiRequest();
  const repoContentRequest = useApiRequest();

  // State
  const availableTeams = ref([]);
  const organizations = ref([]);
  const repositories = ref([]);
  const pullRequests = ref([]);
  const selectedOrg = ref(null);
  const selectedRepo = ref(null);
  const loading = computed(() => 
    organizationsRequest.loading.value || 
    repositoriesRequest.loading.value || 
    pullRequestsRequest.loading.value ||
    repoContentRequest.loading.value
  );
  const error = ref(null);

  // Watch for errors from all requests
  watch([
    organizationsRequest.error, 
    repositoriesRequest.error, 
    pullRequestsRequest.error,
    repoContentRequest.error
  ], ([orgError, repoError, pullError, contentError]) => {
    // Set the first non-null error
    error.value = orgError || repoError || pullError || contentError;
  });

  // Getters remain the same
  const orgOptions = computed(() => {
    return organizations.value.map(org => ({
      value: org.login,
      label: org.name || org.login,
      avatar: org.avatar_url,
      isPersonal: org.isPersonal || false
    }));
  });

  const getAvailableTeams = computed(() => {
    return availableTeams.value;
  });

  const repoOptions = computed(() => {
    return repositories.value.map(repo => ({
      value: repo.full_name,
      label: repo.name,
      isPrivate: repo.private
    }));
  });

  const currentRepo = computed(() => {
    if (!selectedRepo.value) return null;
    return repositories.value.find(repo => repo.full_name === selectedRepo.value);
  });

  const currentOrg = computed(() => {
    if (!selectedOrg.value) return null;
    return organizations.value.find(org => org.login === selectedOrg.value);
  });

  const isPersonalAccount = computed(() => {
    if (!currentOrg.value) return false;
    return currentOrg.value.isPersonal === true;
  });

  // Actions - refactored to use the composable
  async function fetchOrganizations() {
    // Execute the API request
    const response = await organizationsRequest.execute(
      () => api.orgs.getUserOrgs(),
      'Failed to load organizations',
      { organizations: [] }
    );
    
    // Process the response
    organizations.value = response.organizations || [];

    // Add the user as a special item for personal repos
    if (authStore.user) {
      organizations.value.unshift({
        id: authStore.user.id,
        login: authStore.user.login,
        name: authStore.user.login,
        avatar_url: authStore.user.avatar_url,
        isPersonal: true // Flag to identify as personal account
      });
    }

    // Reset selected org if it no longer exists in the list
    if (selectedOrg.value && !organizations.value.some(org => org.login === selectedOrg.value)) {
      selectedOrg.value = null;
    }

    // Default to personal repos if none selected
    if (!selectedOrg.value && organizations.value.length > 0) {
      selectedOrg.value = organizations.value[0].login;
      await fetchRepositories();
    }
  }

  function setAvailableTeams(teams) {
    availableTeams.value = teams;
  }

  async function fetchRepositories() {
    // Reset selected repo before loading new repos
    selectedRepo.value = null;
    
    // Determine which API endpoint to use based on selection
    const orgData = organizations.value.find(org => org.login === selectedOrg.value);
    
    let apiCall;
    if (orgData && orgData.isPersonal) {
      // Use the user repositories endpoint for personal repos
      apiCall = () => api.repos.getUserReposByUsername(orgData.login);
    } else if (selectedOrg.value) {
      // Use the organization repositories endpoint
      apiCall = () => api.repos.getOrgRepos(selectedOrg.value);
    } else {
      // Fallback to current user's repos
      apiCall = () => api.repos.getUserRepos();
    }

    // Execute the API request
    const response = await repositoriesRequest.execute(
      apiCall,
      'Failed to load repositories',
      { repositories: [] }
    );
    
    repositories.value = response.repositories || [];
  }

  async function fetchPullRequests() {
    const response = await pullRequestsRequest.execute(
      () => api.repos.getPullRequests(),
      'Failed to load pull requests',
      { pull_requests: [] }
    );
    
    pullRequests.value = response.pull_requests || [];
  }

  function selectOrganization(orgLogin) {
    selectedOrg.value = orgLogin;
    selectedRepo.value = null;
    fetchRepositories();
  }

  function selectRepository(repoFullName) {
    selectedRepo.value = repoFullName;
  }

  async function fetchRepoContent(path = '', ref = '') {
    if (!currentRepo.value) return null;

    const [owner, repo] = currentRepo.value.full_name.split('/');
    
    const response = await repoContentRequest.execute(
      () => api.repos.getRepoContent(owner, repo, path, ref),
      'Failed to load repository content',
      null
    );

    // Sanitize content if it's text
    if (response && response.content && response.type === 'file') {
      response.content = DOMPurify.sanitize(response.content);
    }

    return response;
  }

  // Reset store state
  function reset() {
    repositories.value = [];
    organizations.value = [];
    pullRequests.value = [];
    selectedOrg.value = null;
    selectedRepo.value = null;
    error.value = null;
  }

  // Return state, getters, and actions
  return {
    // State
    availableTeams,
    repositories,
    organizations,
    pullRequests,
    selectedOrg,
    selectedRepo,
    loading,
    error,

    // Getters
    orgOptions,
    repoOptions,
    getAvailableTeams,
    currentRepo,
    currentOrg,
    isPersonalAccount,

    // Actions
    fetchOrganizations,
    fetchRepositories,
    fetchPullRequests,
    selectOrganization,
    selectRepository,
    fetchRepoContent,
    setAvailableTeams,
    reset
  };
});