// src/stores/github.js
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
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
  const branchesRequest = useApiRequest(); // New request composable for branches

  // State
  const availableTeams = ref([]);
  const repositories = ref([]);
  const organizations = ref([]);
  const pullRequests = ref([]);
  const branches = ref([]); // New state for branches
  const selectedOrg = ref(null);
  const selectedRepo = ref(null);
  const loading = computed(() => 
    organizationsRequest.loading.value || 
    repositoriesRequest.loading.value || 
    pullRequestsRequest.loading.value ||
    repoContentRequest.loading.value ||
    branchesRequest.loading.value // Add branches loading state
  );
  const error = ref(null);

  // Add cache timestamp tracking
  const lastFetchTime = ref({
    organizations: null,
    repositories: null,
    pullRequests: null,
    repoContent: {},  // Will store path-specific timestamps
    branches: null    // Add branches timestamp
  });
  
  // Cache duration (5 minutes)
  const CACHE_DURATION = 5 * 60 * 1000;

  // Cache validation helper
  const isCacheValid = (cacheKey) => {
    return lastFetchTime.value[cacheKey] && 
           (Date.now() - lastFetchTime.value[cacheKey] < CACHE_DURATION);
  };

  // Force refresh helper (for when we need to bypass cache)
  const invalidateCache = (cacheKey) => {
    if (cacheKey) {
      lastFetchTime.value[cacheKey] = null;
    } else {
      // Invalidate all caches
      Object.keys(lastFetchTime.value).forEach(key => {
        if (typeof lastFetchTime.value[key] === 'object') {
          lastFetchTime.value[key] = {};
        } else {
          lastFetchTime.value[key] = null;
        }
      });
    }
  };

  // Getters
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

  // New getter for branch options
  const branchOptions = computed(() => {
    return branches.value.map(branch => ({
      value: branch.name,
      label: branch.name,
      isDefault: branch.is_default
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
  async function fetchOrganizations(forceRefresh = false) {
    // Skip if data is fresh and not forcing refresh
    if (!forceRefresh && organizations.value.length > 0 && isCacheValid('organizations')) {
      console.log('Using cached organizations data');
      return organizations.value;
    }
    
    // Execute the API request
    const response = await organizationsRequest.execute(
      () => api.orgs.getUserOrgs(),
      {
        defaultErrorMsg: 'Failed to load organizations',
        defaultValue: { organizations: [] }
      }
    );
    
    // Update cache timestamp
    lastFetchTime.value.organizations = Date.now();
    
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
    
    return organizations.value;
  }

  function setAvailableTeams(teams) {
    availableTeams.value = teams;
  }

  async function fetchRepositories(forceRefresh = false) {
    // Skip if data is fresh and not forcing refresh
    // Only use cache if we're fetching the same org
    if (
      !forceRefresh && 
      repositories.value.length > 0 && 
      isCacheValid('repositories') &&
      // The following ensures we're not using cached repos from a different org
      lastFetchedOrg.value === selectedOrg.value
    ) {
      console.log('Using cached repositories data');
      return repositories.value;
    }
    
    // Reset selected repo before loading new repos
    selectedRepo.value = null;
    branches.value = []; // Clear branches when fetching new repositories
    
    // Track which org these repos belong to (for cache validation)
    lastFetchedOrg.value = selectedOrg.value;
    
    // Determine which API endpoint to use based on selection
    const orgData = organizations.value.find(org => org.login === selectedOrg.value);
    
    let apiCall;
    if (orgData && orgData.isPersonal) {
      // Use the user repositories endpoint for personal repos
      // apiCall = () => api.repos.getUserReposByUsername(orgData.login);
      apiCall = () => api.repos.getUserRepos();
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
      {
        defaultErrorMsg: 'Failed to load repositories',
        defaultValue: { repositories: [] }
      }
    );
    
    // Update cache timestamp
    lastFetchTime.value.repositories = Date.now();
    
    repositories.value = response.repositories || [];
    
    return repositories.value;
  }

  // New action to fetch branches for a repository
  async function fetchBranches(forceRefresh = false) {
    // Skip if no repository selected
    if (!currentRepo.value) {
      console.log('No repository selected, clearing branches');
      branches.value = [];
      return [];
    }
    
    // Skip if data is fresh and not forcing refresh
    if (
      !forceRefresh && 
      branches.value.length > 0 && 
      isCacheValid('branches') &&
      lastFetchedRepo.value === selectedRepo.value
    ) {
      console.log('Using cached branches data');
      return branches.value;
    }
    
    // Track which repo these branches belong to (for cache validation)
    lastFetchedRepo.value = selectedRepo.value;
    
    const [owner, repo] = currentRepo.value.full_name.split('/');
    console.log(`Fetching branches for ${owner}/${repo}`);
    
    try {
      // Execute the API request
      const response = await branchesRequest.execute(
        () => api.repos.getRepoBranches(owner, repo),
        {
          defaultErrorMsg: 'Failed to load repository branches',
          defaultValue: { branches: [] }
        }
      );
      
      console.log('Branches API response:', response);
      
      if (!response || !Array.isArray(response.branches)) {
        console.error('Invalid branches response format:', response);
        throw new Error('Invalid branches response format');
      }
      
      // Update cache timestamp
      lastFetchTime.value.branches = Date.now();
      
      branches.value = response.branches;
      console.log(`Loaded ${branches.value.length} branches`);
      
      return branches.value;
    } catch (error) {
      console.error('Error fetching branches:', error);
      error.value = error.message || 'Failed to load repository branches';
      branches.value = [];
      throw error;
    }
  }

  async function fetchPullRequests(forceRefresh = false) {
    // Skip if data is fresh and not forcing refresh
    if (!forceRefresh && pullRequests.value.length > 0 && isCacheValid('pullRequests')) {
      console.log('Using cached pull requests data');
      return pullRequests.value;
    }
    
    const response = await pullRequestsRequest.execute(
      () => api.repos.getPullRequests(),
      {
        defaultErrorMsg: 'Failed to load pull requests',
        defaultValue: { pull_requests: [] }
      }
    );
    
    // Update cache timestamp
    lastFetchTime.value.pullRequests = Date.now();
    
    pullRequests.value = response.pull_requests || [];
    
    return pullRequests.value;
  }

  function selectOrganization(orgLogin) {
    if (selectedOrg.value === orgLogin) {
      return; // No change needed
    }
    
    selectedOrg.value = orgLogin;
    selectedRepo.value = null;
    branches.value = []; // Clear branches when changing organization
    
    // We changed org, so fetch repositories
    fetchRepositories();
  }

  async function selectRepository(repoFullName) {
    if (selectedRepo.value === repoFullName) {
      return; // No change needed
    }
    
    console.log(`Selecting repository: ${repoFullName}`);
    selectedRepo.value = repoFullName;
    branches.value = []; // Clear branches when changing repository
    
    if (repoFullName) {
      try {
        // Fetch branches for the selected repository and await the result
        await fetchBranches(true); // Force refresh when explicitly selecting a repo
      } catch (error) {
        console.error('Failed to fetch branches during repository selection:', error);
        // Don't throw here to prevent the UI from breaking if branch fetch fails
      }
    }
  }

  async function fetchRepoContent(path = '', ref = '') {
    if (!currentRepo.value) return null;

    const [owner, repo] = currentRepo.value.full_name.split('/');
    
    // Create a cache key for this specific content
    const cacheKey = `${owner}/${repo}/${path}/${ref}`;
    
    // Skip if data is fresh for this path
    if (
      lastFetchTime.value.repoContent[cacheKey] &&
      (Date.now() - lastFetchTime.value.repoContent[cacheKey] < CACHE_DURATION)
    ) {
      console.log(`Using cached repository content for: ${cacheKey}`);
      if (repoContentCache.value[cacheKey]) {
        return repoContentCache.value[cacheKey];
      }
    }
    
    const response = await repoContentRequest.execute(
      () => api.repos.getRepoContent(owner, repo, path, ref),
      {
        defaultErrorMsg: 'Failed to load repository content',
        defaultValue: null
      }
    );

    // Update cache timestamp for this path
    lastFetchTime.value.repoContent[cacheKey] = Date.now();
    
    // Store response in path-specific cache
    if (response) {
      // Sanitize content if it's text
      if (response.content && response.type === 'file') {
        response.content = DOMPurify.sanitize(response.content);
      }
      
      // Store in cache
      repoContentCache.value[cacheKey] = response;
    }

    return response;
  }

  async function registerService(repo, teamName, prodBranch, devBranch) {
    if (!repo || !teamName) return null;
    
    // Add is_enterprise field if it doesn't exist
    if (repo.is_enterprise === undefined) {
      repo.is_enterprise = false;
    }
  
    try {
      // Show loading state
      const response = await api.services.sendServiceRepoInfo(
        repo,
        teamName,
        prodBranch,
        devBranch
      );
      
      return response.data;
    } catch (error) {
      console.error('Failed to register service:', error);
      throw error;
    }
  }

  // Add a content cache
  const repoContentCache = ref({});
  
  // Track the last org and repo we fetched for
  const lastFetchedOrg = ref(null);
  const lastFetchedRepo = ref(null);

  // Reset store state with cache invalidation
  function reset() {
    repositories.value = [];
    organizations.value = [];
    pullRequests.value = [];
    branches.value = [];
    selectedOrg.value = null;
    selectedRepo.value = null;
    error.value = null;
    
    // Reset all caches
    invalidateCache();
  }

  // Return state, getters, and actions
  return {
    // State
    availableTeams,
    repositories,
    organizations,
    pullRequests,
    branches,
    selectedOrg,
    selectedRepo,
    loading,
    error,

    // Getters
    orgOptions,
    repoOptions,
    branchOptions,
    getAvailableTeams,
    currentRepo,
    currentOrg,
    isPersonalAccount,

    // Actions
    fetchOrganizations,
    fetchRepositories,
    fetchPullRequests,
    fetchBranches,
    selectOrganization,
    selectRepository,
    fetchRepoContent,
    registerService,
    setAvailableTeams,
    reset,
    
    // Cache management
    invalidateCache
  };
});