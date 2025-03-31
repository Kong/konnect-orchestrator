// API service for communicating with the backend
import axios from 'axios';
import authApi from './auth';
import userApi from './user';
import reposApi from './repositories';
import orgsApi from './organizations';
import platformApi from './platform';

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

export default {
    auth: authApi(apiClient),
    user: userApi(apiClient),
    repos: reposApi(apiClient),
    orgs: orgsApi(apiClient),
    platform: platformApi(apiClient),
    
    // Also export the token management functions directly
    setCsrfToken,
    getCsrfToken
  };