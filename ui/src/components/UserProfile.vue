<template>
    <div v-if="authStore.isAuthenticated && authStore.user" class="user-profile">
      <div class="user-profile-header">
        <img :src="authStore.avatar" alt="User avatar" class="avatar" />
        <div class="user-info">
          <h2 class="user-name">{{ authStore.user.name || authStore.user.login }}</h2>
          <p class="user-login">@{{ authStore.user.login }}</p>
          <p v-if="authStore.user.email" class="user-email">{{ authStore.user.email }}</p>
        </div>
      </div>
      
      <div v-if="authStore.user.bio" class="user-bio">
        {{ authStore.user.bio }}
      </div>
      
      <div class="user-meta">
        <div v-if="authStore.user.company" class="meta-item">
          <span class="meta-icon">🏢</span>
          <span>{{ authStore.user.company }}</span>
        </div>
        <div v-if="authStore.user.location" class="meta-item">
          <span class="meta-icon">📍</span>
          <span>{{ authStore.user.location }}</span>
        </div>
      </div>
      
      <div class="user-actions">
        <a :href="authStore.user.html_url" target="_blank" rel="noopener noreferrer" class="github-link">
          View GitHub Profile
        </a>
        <button @click="logout" class="logout-button">Logout</button>
      </div>
    </div>
  </template>
  
  <script setup>
  import { useAuthStore } from '@/stores/auth';
  
  const authStore = useAuthStore();
  
  // Methods
  const logout = () => {
    authStore.logout();
  };
  </script>
  
  <style scoped>
  .user-profile {
    width: 100%;
    max-width: 500px;
    padding: 1.5rem;
    border: 1px solid #e1e4e8;
    border-radius: 6px;
    background-color: #f6f8fa;
    margin: 1rem 0;
  }
  
  .user-profile-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
  }
  
  .avatar {
    width: 60px;
    height: 60px;
    border-radius: 50%;
    border: 1px solid #e1e4e8;
  }
  
  .user-info {
    flex: 1;
  }
  
  .user-name {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }
  
  .user-login, .user-email {
    margin: 4px 0 0 0;
    color: #586069;
    font-size: 14px;
  }
  
  .user-bio {
    margin: 1rem 0;
    font-size: 14px;
    line-height: 1.5;
  }
  
  .user-meta {
    margin: 1rem 0;
  }
  
  .meta-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
    font-size: 14px;
    color: #586069;
  }
  
  .meta-icon {
    font-size: 16px;
  }
  
  .user-actions {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-top: 1.5rem;
  }
  
  .github-link {
    display: inline-block;
    padding: 6px 12px;
    background-color: #0366d6;
    color: white;
    border-radius: 6px;
    text-decoration: none;
    font-size: 14px;
    font-weight: 500;
    transition: background-color 0.2s;
  }
  
  .github-link:hover {
    background-color: #0250a4;
  }
  
  .logout-button {
    padding: 6px 12px;
    background-color: transparent;
    color: #24292e;
    border: 1px solid #d1d5da;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .logout-button:hover {
    background-color: #f3f4f6;
  }
  </style>