// API service for communicating with the backend
import axios from 'axios';

// Create axios instance with base URL from env
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
});

// In api.js or where you handle CSRF tokens:
let csrfToken = '';

// When setting the token, also store it in sessionStorage
export function setCsrfToken(token) {
  csrfToken = token;
  if (token) {
    sessionStorage.setItem('csrf_token', token);
  } else {
    sessionStorage.removeItem('csrf_token');
  }
}

// When getting the token, try to recover it from sessionStorage
export function getCsrfToken() {
  if (!csrfToken && typeof sessionStorage !== 'undefined') {
    csrfToken = sessionStorage.getItem('csrf_token') || '';
  }
  return csrfToken;
}

// Add interceptor to include the auth token in requests
apiClient.interceptors.request.use(config => {
  // Add Authorization header if we have a token in localStorage
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  // Add CSRF token for non-GET requests
  if (config.method !== 'get' && csrfToken) {
    config.headers['X-CSRF-Token'] = csrfToken;
  }

  return config;
});

apiClient.interceptors.response.use(
  response => {
    return response;
  },
  async error => {
    // Don't redirect if the error is from the logout endpoint
    const isLogoutRequest = error.config && error.config.url === '/auth/logout';
    
    // Don't redirect if already on home page or auth pages
    const isOnPublicPage = window.location.pathname === '/' || 
                          window.location.pathname.startsWith('/auth');
    
    // Handle 401 errors (unauthorized)
    if (error.response && error.response.status === 401 && !isLogoutRequest) {
      // Update the auth store to prevent repeated auth attempts
      const authStore = useAuthStore();
      authStore.authenticationFailed = true;
      
      // Clear token 
      localStorage.removeItem('auth_token');
      setCsrfToken('');
      
      // Only redirect if not already on a public page
      if (!isOnPublicPage) {
        window.location.href = '/';
      }
    }
    return Promise.reject(error);
  }
);

// API methods
const api = {
  // Auth endpoints
  auth: {
    // Get GitHub login URL
    getLoginUrl() {
      return `${apiClient.defaults.baseURL}/auth/github`;
    },

    verifyCode(code) {
      return apiClient.get('/auth/verify', { params: { code } });
    },

    // Exchange code for token (if implementing client-side)
    // This is typically handled by the server-side callback
    exchangeToken(code) {
      return apiClient.post('/auth/exchange', { code });
    },

    // Handle logout
    logout(config = {}) {
      return apiClient.post('/auth/logout', {}, config);
    },

    // Refresh token
    refreshToken() {
      return apiClient.post('/auth/refresh');
    }
  },

  // User endpoints
  user: {
    // Get current user profile
    getProfile() {
      return apiClient.get('/api/user');
    }
  },

  // Organization endpoints
  orgs: {
    // Get user's organizations
    getUserOrgs() {
      return apiClient.get('/api/orgs');
    }
  },

  // Repository endpoints
  repos: {
    /**
  * Get authenticated user's repositories
  * @returns {Promise<Object>} Response containing user's repositories
  */
    getUserRepos() {
      return apiClient.get('/api/repos', {
        params: { visibility: 'all' }
      });
    },

    /**
     * Get repositories for a specific user
     * @param {string} username - GitHub username
     * @returns {Promise<Object>} Response containing user's repositories
     */
    getUserReposByUsername(username) {
      return apiClient.get(`/api/users/${username}/repos`, {
        params: { visibility: 'all' }
      });
    },

    /**
     * Get repositories for an organization
     * @param {string} org - Organization name
     * @returns {Promise<Object>} Response containing organization's repositories
     */
    getOrgRepos(org) {
      return apiClient.get(`/api/orgs/${org}/repos`, {
        params: { visibility: 'all' }
      });
    },

    getRepoBranches(owner, repo) {
      return apiClient.get(`/api/repos/${owner}/${repo}/branches`);
    },

    // Get repository content
    getRepoContent(owner, repo, path = '', ref = '') {
      return apiClient.get(`/api/repos/${owner}/${repo}/contents/${path}${ref ? `?ref=${ref}` : ''}`);
    },

    // Get pull requests for a repository
    getPullRequests(state = 'all') {
      return apiClient.get(`/api/platform/pulls`, {
        params: { state }
      });
    }

  },

  // Services endpoints
  services: {
    // Get all services
    getServices() {
      return apiClient.get('/api/platform/service');
    },

    sendServiceRepoInfo(repo, team, prodBranch, devBranch) {
      // Create a new object that matches the exact structure expected by the backend
      const repoData = {
        id: repo.id,
        name: repo.name,
        full_name: repo.full_name,
        description: repo.description || '',
        private: repo.private,
        html_url: repo.html_url,
        clone_url: repo.clone_url || '',
        ssh_url: repo.ssh_url || '',
        owner: {
          login: repo.owner.login,
          id: repo.owner.id,
          avatar_url: repo.owner.avatar_url
        },
        default_branch: repo.default_branch || '',
        created_at: repo.created_at || '',
        updated_at: repo.updated_at || '',
        is_enterprise: repo.is_enterprise || false,
        team: team || '',
        prodBranch: prodBranch,
        devBranch: devBranch
      };

      return apiClient.post('/api/platform/service', repoData);
    }
  }
};

export default api;