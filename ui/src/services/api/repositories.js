// src/services/api/reposApi.js

/**
 * Repositories API endpoints
 * @param {Object} apiClient - Axios instance
 * @returns {Object} Repositories API methods
 */
export default function reposApi(apiClient) {
  return {
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
    
    /**
     * Get content from a repository
     * @param {string} owner - Repository owner
     * @param {string} repo - Repository name
     * @param {string} [path=''] - File or directory path
     * @param {string} [ref=''] - Branch or commit reference
     * @returns {Promise<Object>} Response containing repository content
     */
    getRepoContent(owner, repo, path = '', ref = '') {
      return apiClient.get(`/api/repos/${owner}/${repo}/contents/${path}${ref ? `?ref=${ref}` : ''}`);
    },
    
    /**
     * Get branches for a repository
     * @param {string} owner - Repository owner
     * @param {string} repo - Repository name
     * @returns {Promise<Object>} Response containing repository branches
     */
    getRepoBranches(owner, repo) {
      return apiClient.get(`/api/repos/${owner}/${repo}/branches`);
    },
    
    /**
     * Get pull requests for a repository
     * @param {string} [state='all'] - Pull request state (all, open, closed)
     * @returns {Promise<Object>} Response with pull requests
     */
    getPullRequests(state = 'all') {
      return apiClient.get(`/api/platform/pulls`, {
        params: { state }
      });
    }
  };
}