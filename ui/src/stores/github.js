// GitHub repositories store using Pinia
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import api from '@/services/api';
import { useAuthStore } from '@/stores/auth';
import DOMPurify from 'dompurify';

export const useGithubStore = defineStore('github', () => {
  // Get auth store for access to user info
  const authStore = useAuthStore();
  
  // State
  const repositories = ref([]);
  const organizations = ref([]);
  const pullRequests = ref([]);
  const selectedOrg = ref(null);
  const selectedRepo = ref(null);
  const loading = ref(false);
  const error = ref(null);
  
  // Getters
  const orgOptions = computed(() => {
    return organizations.value.map(org => ({
      value: org.login,
      label: org.name || org.login,
      avatar: org.avatar_url,
      isPersonal: org.isPersonal || false
    }));
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
  
  // Actions
  async function fetchOrganizations() {
    loading.value = true;
    error.value = null;
    
    try {
      // Get organizations
      const response = await api.orgs.getUserOrgs();
      organizations.value = response.data.organizations || [];
      
      // Add the user as a special item for personal repos
      if (authStore.user) {
        organizations.value.unshift({
          id: authStore.user.id,
          login: authStore.user.login,
          name: `${authStore.user.login} (Personal)`,
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
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to load organizations';
      organizations.value = [];
    } finally {
      loading.value = false;
    }
  }
  
  async function fetchRepositories() {
    loading.value = true;
    error.value = null;
    selectedRepo.value = null;
    
    try {
      let response;
      
      // Determine which API endpoint to use based on selection
      const orgData = organizations.value.find(org => org.login === selectedOrg.value);
      
      if (orgData && orgData.isPersonal) {
        // Use the user repositories endpoint for personal repos
        response = await api.repos.getUserReposByUsername(orgData.login);
      } else if (selectedOrg.value) {
        // Use the organization repositories endpoint
        response = await api.repos.getOrgRepos(selectedOrg.value);
      } else {
        // Fallback to current user's repos
        response = await api.repos.getUserRepos();
      }
      
      repositories.value = response.data.repositories || [];
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to load repositories';
      repositories.value = [];
    } finally {
      loading.value = false;
    }
  }
  
  async function fetchPullRequests() {
   
    loading.value = true;
    error.value = null;
    
    try {
      const response = await api.repos.getPullRequests();
      pullRequests.value = response.data.pull_requests || [];
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to load pull requests';
      pullRequests.value = [];
    } finally {
      loading.value = false;
    }
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
    
    loading.value = true;
    error.value = null;
    
    try {
      const [owner, repo] = currentRepo.value.full_name.split('/');
      const response = await api.repos.getRepoContent(owner, repo, path, ref);
      
      // Sanitize content if it's text
      if (response.data.content && response.data.type === 'file') {
        response.data.content = DOMPurify.sanitize(response.data.content);
      }
      
      return response.data;
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to load repository content';
      return null;
    } finally {
      loading.value = false;
    }
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
    reset
  };
});