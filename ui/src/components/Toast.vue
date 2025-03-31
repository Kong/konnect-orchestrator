<template>
    <div class="toast-container" aria-live="polite" aria-atomic="true">
      <TransitionGroup name="toast">
        <div 
          v-for="toast in toasts" 
          :key="toast.id"
          :class="['toast', `toast-${toast.type}`]"
          @click="remove(toast.id)"
        >
          <div class="toast-icon">
            <svg v-if="toast.type === 'success'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
              <polyline points="22 4 12 14.01 9 11.01"></polyline>
            </svg>
            <svg v-else-if="toast.type === 'error'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="15" y1="9" x2="9" y2="15"></line>
              <line x1="9" y1="9" x2="15" y2="15"></line>
            </svg>
            <svg v-else-if="toast.type === 'info'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="16" x2="12" y2="12"></line>
              <line x1="12" y1="8" x2="12.01" y2="8"></line>
            </svg>
            <svg v-else-if="toast.type === 'warning'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
              <line x1="12" y1="9" x2="12" y2="13"></line>
              <line x1="12" y1="17" x2="12.01" y2="17"></line>
            </svg>
          </div>
          <div class="toast-message">{{ toast.message }}</div>
        </div>
      </TransitionGroup>
    </div>
  </template>
  
  <script setup>
  import { useToast } from '@/composables/useToast';
  
  const { toasts, remove } = useToast();
  </script>
  
  <style scoped>
  .toast-container {
    position: fixed;
    top: 1rem;
    right: 1rem;
    z-index: 1050;
    display: flex;
    flex-direction: column;
    max-width: 350px;
  }
  
  .toast {
    display: flex;
    align-items: center;
    padding: 0.75rem 1rem;
    margin-bottom: 0.5rem;
    border-radius: 8px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    cursor: pointer;
    transition: transform 0.3s ease, opacity 0.3s ease;
    background-color: #1a1a1a;
    border: 1px solid transparent;
  }
  
  .toast-success {
    border-color: #40c057;
    color: #40c057;
  }
  
  .toast-error {
    border-color: #fa5252;
    color: #fa5252;
  }
  
  .toast-info {
    border-color: #228be6;
    color: #228be6;
  }
  
  .toast-warning {
    border-color: #fab005;
    color: #fab005;
  }
  
  .toast-icon {
    margin-right: 0.75rem;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .toast-message {
    flex: 1;
    font-weight: 500;
  }
  
  /* Transitions */
  .toast-enter-active,
  .toast-leave-active {
    transition: all 0.3s ease;
  }
  
  .toast-enter-from {
    opacity: 0;
    transform: translateX(30px);
  }
  
  .toast-leave-to {
    opacity: 0;
    transform: translateX(30px);
  }
  
  @media (prefers-color-scheme: light) {
    .toast {
      background-color: #ffffff;
    }
  }
  </style>