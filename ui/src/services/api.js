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

// CSRF token storage
let csrfToken = '';

// Set CSRF token
export function setCsrfToken(token) {
  csrfToken = token;
}

// Get CSRF token
export function getCsrfToken() {
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

// Add interceptor to handle token refresh
apiClient.interceptors.response.use(
  response => {
    // Check if we got a refresh token and store it
    const refreshToken = response.headers['x-refresh-token'];
    if (refreshToken) {
      localStorage.setItem('auth_token', refreshToken);
    }
    return response;
  },
  async error => {
    // Handle 401 errors (unauthorized)
    if (error.response && error.response.status === 401) {
      // Clear token and redirect to login
      localStorage.removeItem('auth_token');
      setCsrfToken(''); // Clear CSRF token
      window.location.href = '/';
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
    
    // Exchange code for token (if implementing client-side)
    // This is typically handled by the server-side callback
    exchangeToken(code) {
      return apiClient.post('/auth/exchange', { code });
    },
    
    // Handle logout
    logout() {
      return apiClient.post('/auth/logout');
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
    // Get authenticated user's repositories
    getUserRepos() {
      return apiClient.get('/api/repos');
    },
    
    // Get specific user's repositories
    getUserReposByUsername(username) {
      return apiClient.get(`/api/users/${username}/repos`);
    },
    
    // Get organization repositories
    getOrgRepos(org) {
      return apiClient.get(`/api/orgs/${org}/repos`);
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
      return apiClient.get('/api/services');
    },
    
    // Get a specific service by name
    getService(name) {
      return apiClient.get(`/api/services/${name}`);
    }
  }
};

export default api;