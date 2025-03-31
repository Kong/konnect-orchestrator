import { ref } from 'vue';

export function createCache(cacheDuration = 5 * 60 * 1000) {
  const cache = ref({});
  const timestamps = ref({});
  
  function get(key) {
    if (isValid(key)) {
      return cache.value[key];
    }
    return null;
  }
  
  function set(key, value) {
    cache.value[key] = value;
    timestamps.value[key] = Date.now();
  }
  
  function isValid(key) {
    return timestamps.value[key] && 
           (Date.now() - timestamps.value[key] < cacheDuration);
  }
  
  function invalidate(key) {
    if (key) {
      delete cache.value[key];
      delete timestamps.value[key];
    } else {
      cache.value = {};
      timestamps.value = {};
    }
  }
  
  return { get, set, isValid, invalidate };
}