// src/services/api/userApi.js

/**
 * User API endpoints
 * @param {Object} apiClient - Axios instance
 * @returns {Object} User API methods
 */
export default function userApi(apiClient) {
    return {
      /**
       * Get current user profile
       * @returns {Promise<Object>} Response containing user profile data
       */
      getProfile() {
        return apiClient.get('/api/user');
      }
    };
  }