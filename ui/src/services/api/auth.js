/**
 * Authentication API endpoints
 * @param {Object} apiClient - Axios instance
 * @returns {Object} Auth API methods
 */
export default function authApi(apiClient) {
    return {
      /**
       * Get GitHub login URL
       * @returns {string} The login URL
       */
      getLoginUrl() {
        return `${apiClient.defaults.baseURL}/auth/github`;
      },
      
      /**
       * Verify authentication code
       * @param {string} code - The authorization code to verify
       * @returns {Promise<Object>} Response with authentication status
       */
      verifyCode(code) {
        return apiClient.get('/auth/verify', { params: { code } });
      },
  
      /**
       * Exchange code for token (client-side implementation)
       * @param {string} code - Authorization code from OAuth
       * @returns {Promise<Object>} Response with token
       */
      exchangeToken(code) {
        return apiClient.post('/auth/exchange', { code });
      },
      
      /**
       * Logout the current user
       * @returns {Promise<Object>} Response with logout status
       */
      logout() {
        return apiClient.post('/auth/logout');
      },
      
      /**
       * Refresh the authentication token
       * @returns {Promise<Object>} Response with new token
       */
      refreshToken() {
        return apiClient.post('/auth/refresh');
      }
    };
  }