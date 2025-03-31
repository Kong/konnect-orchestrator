// src/services/api/orgsApi.js

/**
 * Organizations API endpoints
 * @param {Object} apiClient - Axios instance
 * @returns {Object} Organizations API methods
 */
export default function orgsApi(apiClient) {
    return {
      /**
       * Get organizations for the authenticated user
       * @returns {Promise<Object>} Response containing user's organizations
       */
      getUserOrgs() {
        return apiClient.get('/api/orgs');
      }
    };
  }