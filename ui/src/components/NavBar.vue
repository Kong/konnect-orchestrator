<template>
    <nav class="navbar">
      <div class="navbar-brand">
        <router-link to="/" class="navbar-logo">
          GitHub Repository Explorer
        </router-link>
      </div>
      
      <!-- For authenticated users -->
      <div v-if="authStore.isAuthenticated" class="navbar-controls">
        <!-- Organization Dropdown -->
        <div class="org-dropdown-container" v-if="$route.path.includes('/dashboard')">
          <OrgDropdown @change="handleOrgChange" />
        </div>
        
        <!-- User profile and logout -->
        <div class="navbar-user">
          <div class="user-info">
            <img :src="authStore.avatar" alt="Avatar" class="user-avatar" />
            <span class="user-name">{{ authStore.username }}</span>
          </div>
          <button @click="logout" class="btn-outline">
            Logout
          </button>
        </div>
      </div>
      
      <!-- For non-authenticated users -->
      <div v-else class="navbar-login">
        <GitHubLogin />
      </div>
    </nav>
  </template>
  
  <script setup>
  import { useRouter, useRoute } from 'vue-router';
  import { useAuthStore } from '@/stores/auth';
  import { useGithubStore } from '@/stores/github';
  import OrgDropdown from '@/components/OrgDropdown.vue';
  import GitHubLogin from '@/components/GitHubLogin.vue';
  
  const router = useRouter();
  const route = useRoute();
  const authStore = useAuthStore();
  const githubStore = useGithubStore();
  
  const logout = async () => {
    await authStore.logout();
    router.push('/');
  };
  
  const handleOrgChange = (orgLogin) => {
    githubStore.selectOrganization(orgLogin);
  };
  </script>
  
  <style scoped>
  .navbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 2rem;
    background-color: var(--color-bg-light);
    border-bottom: 1px solid var(--color-border);
    color: var(--color-text-white);
    width: 100%;
    box-sizing: border-box;
  }
  
  .navbar-logo {
    color: var(--color-text-white);
    font-size: 1.2rem;
    font-weight: 600;
    text-decoration: none;
    display: flex;
    align-items: center;
  }
  
  .navbar-logo::before {
    content: "";
    display: inline-block;
    width: 24px;
    height: 24px;
    margin-right: 10px;
    background-color: var(--color-primary);
    mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3E%3Cpath fill='%23fff' d='M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z'/%3E%3C/svg%3E");
    -webkit-mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3E%3Cpath fill='%23fff' d='M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z'/%3E%3C/svg%3E");
    mask-size: contain;
    -webkit-mask-size: contain;
    mask-repeat: no-repeat;
    -webkit-mask-repeat: no-repeat;
  }
  
  .navbar-controls {
    display: flex;
    align-items: center;
    gap: 2rem;
    flex-wrap: wrap;
  }
  
  .org-dropdown-container {
    min-width: 250px;
    flex-grow: 1; /* Allow the dropdown to grow as needed */
    max-width: 500px; /* Set a reasonable maximum width */
  }
  
  /* Override OrgDropdown styles to fit in navbar */
  .org-dropdown-container :deep(.org-dropdown) {
    margin: 0;
    display: flex;
    align-items: center;
    flex-wrap: nowrap;
  }
  
  /* Make label and dropdown inline */
  .org-dropdown-container :deep(label) {
    color: var(--color-text-light);
    margin-right: 10px;
    margin-bottom: 0;
    display: inline-block;
    white-space: nowrap; /* Prevent label from wrapping */
  }
  
  .org-dropdown-container :deep(.select-container) {
    display: inline-flex;
    width: auto;
    flex: 1;
  }
  
  .org-dropdown-container :deep(select) {
  width: 100%; /* Make select element use full width of container */
  background-color: var(--color-bg-medium);
  color: var(--color-text-white);
  border-color: var(--color-border);
  text-overflow: ellipsis; /* Add ellipsis for very long text */
  padding: 10px 36px 10px 12px; /* Increased right padding to prevent text overlap */
}
  
  .org-dropdown-container :deep(select:hover) {
    border-color: var(--color-primary);
  }
  
  .org-dropdown-container :deep(.error-message),
  .org-dropdown-container :deep(.empty-message) {
    color: #f85149;
  }
  
  .navbar-user {
    display: flex;
    align-items: center;
    gap: 1rem;
  }
  
  .user-info {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  
  .user-avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    border: 1px solid var(--color-border);
  }
  
  .user-name {
    font-size: 14px;
    color: var(--color-text-white);
  }
  
  .logout-button {
    padding: 6px 12px;
    background-color: transparent;
    color: white;
    border: 1px solid var(--color-border);
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .logout-button:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }
  
  /* Style for login button in navbar */
  .navbar-login :deep(.github-button) {
    font-size: 14px;
    padding: 8px 16px;
    background: linear-gradient(90deg, var(--color-accent-blue), var(--color-primary));
    border: none;
    color: white;
    border-radius: 4px;
    transition: all 0.3s;
  }
  
  .navbar-login :deep(.github-button:hover) {
    opacity: 0.9;
    transform: translateY(-1px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
  }
  
  /* Responsive styling */
  @media (max-width: 1200px) {
    .org-dropdown-container {
      min-width: 200px;
    }
    
    .user-name {
      max-width: 100px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
  
  @media (max-width: 992px) {
    .navbar {
      flex-direction: row;
      flex-wrap: wrap;
      gap: 1rem;
      padding: 1rem;
      justify-content: space-between;
    }
    
    .navbar-controls {
      order: 3;
      width: 100%;
      justify-content: space-between;
      margin-top: 0.5rem;
    }
    
    .org-dropdown-container {
      width: 100%;
      max-width: 100%;
    }
    
    .org-dropdown-container :deep(.org-dropdown) {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 0.5rem;
    }
    
    .org-dropdown-container :deep(.select-container) {
      flex: 1;
      min-width: 150px;
    }
  }
  
  @media (max-width: 600px) {
    .navbar-user {
      width: 100%;
      justify-content: space-between;
      margin-top: 0.5rem;
    }
    
    .org-dropdown-container :deep(label) {
      width: 100%;
      margin-bottom: 0.5rem;
    }
    
    .org-dropdown-container :deep(.org-dropdown) {
      flex-direction: column;
      align-items: flex-start;
    }
  }
  </style>