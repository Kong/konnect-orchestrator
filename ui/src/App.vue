<template>
  <div class="app">
    <header class="app-header">
      <div class="header-content">
        <div class="header-left">
          <router-link to="/" class="logo">
            GitHub Explorer
          </router-link>
          
          <nav v-if="authStore.isAuthenticated" class="nav">
            <router-link to="/dashboard" class="nav-link">Dashboard</router-link>
          </nav>
        </div>
        
        <div v-if="authStore.isAuthenticated" class="header-right">
          <!-- Organization Dropdown -->
          <OrgDropdown v-if="authStore.isAuthenticated" class="header-org-dropdown" />
          
          <!-- User Profile Mini -->
          <div class="user-mini" v-if="authStore.isAuthenticated && authStore.user">
            <span class="user-name">{{ authStore.user.name || authStore.user.login }}</span>
            <img :src="authStore.avatar" alt="User avatar" class="avatar-mini" />
          </div>
        </div>
      </div>
    </header>
    
    <main class="app-content">
      <router-view />
    </main>
    
    <footer class="app-footer">
      <div class="footer-content">
        <p>GitHub Explorer © {{ currentYear }}</p>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue';
import { useAuthStore } from '@/stores/auth';
import OrgDropdown from '@/components/OrgDropdown.vue';

const authStore = useAuthStore();

// Computed properties
const currentYear = computed(() => new Date().getFullYear());

// Check for existing token on app mount
onMounted(() => {
  authStore.checkAuth();
});
</script>

<style>
/* Reset and base styles */
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
  font-size: 16px;
  line-height: 1.5;
  color: #24292e;
  background-color: #fff;
}

.app {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

/* Header styles */
.app-header {
  background-color: #24292e;
  color: #fff;
  padding: 1rem 0;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.logo {
  font-size: 1.25rem;
  font-weight: 600;
  color: white;
  text-decoration: none;
  margin-right: 2rem;
}

.nav {
  display: flex;
  gap: 1.5rem;
}

.nav-link {
  color: #c8c9cb;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}

.nav-link:hover,
.nav-link.router-link-active {
  color: white;
}

/* User mini profile in header */
.user-mini {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.user-name {
  font-size: 0.9rem;
  font-weight: 500;
  color: #fff;
}

.avatar-mini {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

/* Header org dropdown styles */
.header-org-dropdown {
  margin: 0;
  max-width: none;
}

.header-org-dropdown :deep(label) {
  display: none;
}

.header-org-dropdown :deep(select) {
  background-color: #2c3137;
  color: white;
  border-color: #444;
  padding: 6px 28px 6px 10px;
  font-size: 0.9rem;
}

.header-org-dropdown :deep(.select-container) {
  min-width: 200px;
}

.header-org-dropdown :deep(.error-message),
.header-org-dropdown :deep(.empty-message),
.header-org-dropdown :deep(.org-info) {
  display: none;
}

/* Main content styles */
.app-content {
  flex: 1;
}

/* Footer styles */
.app-footer {
  background-color: #f6f8fa;
  border-top: 1px solid #e1e4e8;
  padding: 1.5rem 0;
  margin-top: 2rem;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
  text-align: center;
  color: #586069;
  font-size: 0.875rem;
}
</style>