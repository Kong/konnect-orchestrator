// src/composables/useToast.js
import { ref, readonly } from 'vue';

// Shared state for toast notifications
const toasts = ref([]);
let nextId = 0;

export function useToast() {
  // Toast creation helper function
  const createToast = (type, message, timeout = 3000) => {
    const id = nextId++;
    const toast = {
      id,
      type,
      message,
      timestamp: Date.now()
    };
    
    toasts.value.push(toast);
    
    // Auto-remove toast after timeout
    if (timeout > 0) {
      setTimeout(() => {
        removeToast(id);
      }, timeout);
    }
    
    return id;
  };
  
  // Remove a toast by id
  const removeToast = (id) => {
    const index = toasts.value.findIndex(toast => toast.id === id);
    if (index !== -1) {
      toasts.value.splice(index, 1);
    }
  };
  
  // Public methods for different toast types
  const success = (message, timeout) => createToast('success', message, timeout);
  const error = (message, timeout) => createToast('error', message, timeout);
  const info = (message, timeout) => createToast('info', message, timeout);
  const warning = (message, timeout) => createToast('warning', message, timeout);
  
  // Clear all toasts
  const clearAll = () => {
    toasts.value = [];
  };
  
  return {
    // Expose read-only toasts list
    toasts: readonly(toasts),
    // Toast creation methods
    success,
    error,
    info,
    warning,
    // Toast management methods
    remove: removeToast,
    clearAll
  };
}