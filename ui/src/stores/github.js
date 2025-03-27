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

  // State
  const availableTeams = ref([]);
  const repositories = ref([]);
  const organizations = ref([]);
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

  // Add cache timestamp tracking
  const lastFetchTime = ref({
    organizations: null,
    repositories: null,
    pullRequests: null,
    repoContent: {}  // Will store path-specific timestamps
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

  // Actions - updated with caching
  async function fetchOrganizations(forceRefresh = false) {
    // Skip if data is fresh and not forcing refresh
    if (!forceRefresh && organizations.value.length > 0 && isCacheValid('organizations')) {
      console.log('Using cached organizations data');
      return organizations.value;
    }
    
    // Execute the API request
    const response = await organizationsRequest.execute(
      () => api.orgs.getUserOrgs(),
      'Failed to load organizations',
      { organizations: [] }
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
    
    // Track which org these repos belong to (for cache validation)
    lastFetchedOrg.value = selectedOrg.value;
    
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
    
    // Update cache timestamp
    lastFetchTime.value.repositories = Date.now();
    
    repositories.value = response.repositories || [];
    
    return repositories.value;
  }

  async function fetchPullRequests(forceRefresh = false) {
    // Skip if data is fresh and not forcing refresh
    if (!forceRefresh && pullRequests.value.length > 0 && isCacheValid('pullRequests')) {
      console.log('Using cached pull requests data');
      return pullRequests.value;
    }
    
    const response = await pullRequestsRequest.execute(
      () => api.repos.getPullRequests(),
      'Failed to load pull requests',
      { pull_requests: [] }
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
    
    // We changed org, so fetch repositories
    fetchRepositories();
  }

  function selectRepository(repoFullName) {
    selectedRepo.value = repoFullName;
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
      'Failed to load repository content',
      null
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

  // Add a content cache
  const repoContentCache = ref({});
  
  // Track the last org we fetched repos for
  const lastFetchedOrg = ref(null);

  // Reset store state with cache invalidation
  function reset() {
    repositories.value = [];
    organizations.value = [];
    pullRequests.value = [];
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
    reset,
    
    // Cache management
    invalidateCache
  };
});