// src/composables/useApiRequest.js
import { ref } from 'vue';

export function useApiRequest() {
  const data = ref(null);
  const loading = ref(false);
  const error = ref(null);

  async function execute(apiCall, defaultErrorMsg = 'An error occurred', defaultValue = null) {
    loading.value = true;
    error.value = null;
    
    try {
      const response = await apiCall();
      data.value = response.data;
      return response.data;
    } catch (err) {
      console.error('API Error:', err);
      error.value = err.response?.data?.error || defaultErrorMsg;
      data.value = defaultValue;
      return defaultValue;
    } finally {
      loading.value = false;
    }
  }

  return {
    data,
    loading,
    error,
    execute
  };
}