// src/services/api/platformApi.js

/**
 * Platform API endpoints for platform-specific operations
 * @param {Object} apiClient - Axios instance
 * @returns {Object} Platform API methods
 */
export default function platformApi(apiClient) {
    return {
      /**
       * Get all platform services
       * @returns {Promise<Object>} Response containing services data
       */
      getServices() {
        return apiClient.get('/api/platform/service');
      },
    
      /**
       * Register or update a service repository
       * @param {Object} repo - Repository object with details
       * @param {string} team - Team name
       * @param {string} [prodBranch='prod'] - Production branch name
       * @param {string} [devBranch='dev'] - Development branch name
       * @returns {Promise<Object>} Response with operation result
       */
      sendServiceRepoInfo(repo, team) {
        // Create a new object that matches the expected backend structure
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
      },
      
      /**
       * Get pull requests for the platform repository
       * @param {string} [state='all'] - Pull request state (all, open, closed)
       * @param {string} [sort='created'] - Sort field
       * @param {string} [direction='desc'] - Sort direction
       * @returns {Promise<Object>} Response with pull requests
       */
      getPullRequests(state = 'all', sort = 'created', direction = 'desc') {
        return apiClient.get('/api/platform/pulls', {
          params: { state, sort, direction }
        });
      }
    };
  }