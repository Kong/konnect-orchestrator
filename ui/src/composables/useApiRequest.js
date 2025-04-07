import { ref } from 'vue';

export function useApiRequest() {
  const data = ref(null);
  const loading = ref(false);
  const error = ref(null);
  const retries = ref(0);
  
  async function execute(
    apiCall, 
    { 
      defaultErrorMsg = 'An error occurred', 
      defaultValue = null,
      maxRetries = 0,
      retryDelay = 1000,
      shouldRetry = (error) => error.response?.status >= 500
    } = {}
  ) {
    loading.value = true;
    error.value = null;
    retries.value = 0;
    
    async function tryRequest() {
      try {
        const response = await apiCall();
        data.value = response.data;
        return response.data;
      } catch (err) {
        // Check if we should retry
        if (retries.value < maxRetries && shouldRetry(err)) {
          retries.value++;
          await new Promise(resolve => setTimeout(resolve, retryDelay));
          return tryRequest();
        }
        
        console.error('API Error:', err);
        
        // Enhanced error details
        error.value = {
          message: err.response?.data?.error || defaultErrorMsg,
          status: err.response?.status,
          type: err.response?.status >= 500 ? 'server' : 'client',
          original: err
        };
        
        data.value = defaultValue;
        return defaultValue;
      }
    }
    
    try {
      return await tryRequest();
    } finally {
      loading.value = false;
    }
  }

  return {
    data,
    loading,
    error,
    retries,
    execute
  };
}