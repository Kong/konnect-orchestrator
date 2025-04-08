<template>
  <nav class="navbar">
    <div class="navbar-brand">
      <router-link to="/" class="navbar-logo">
        {{ title }}
      </router-link>
    </div>
    
    <!-- For authenticated users -->
    <div v-if="authStore.isAuthenticated" class="navbar-controls">
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
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import GitHubLogin from '@/components/GitHubLogin.vue';

const router = useRouter();
const authStore = useAuthStore();
const title = import.meta.env.VITE_APP_TITLE
const logout = async () => {
  await authStore.logout();
  router.push('/');
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
  mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 29 26'%3E%3Cpath clip-rule='evenodd' d='m9.144 21.802.693-.894h5.126l2.672 3.415-.475 1.159h-6.61l.158-1.159-1.564-2.52Z' fill='%23169FCC' fill-rule='evenodd'%3E%3C/path%3E%3Cpath clip-rule='evenodd' d='m10.711 10.158 2.476-4.278h2.884L29 20.903l-1.003 4.579h-5.543l.347-1.286-12.09-14.038Z' fill='%2314B59A' fill-rule='evenodd'%3E%3C/path%3E%3Cpath clip-rule='evenodd' d='m13.717 5.068 1.185-2.193L18.457 0l6.093 4.778-.79.807 1.06 1.468v1.572l-3.036 2.482-5.094-6.04h-2.973Z' fill='%231BC263' fill-rule='evenodd'%3E%3C/path%3E%3Cpath clip-rule='evenodd' d='M4.301 14.621h1.61l4.2-3.514 5.565 6.469-1.57 2.356H8.97l-3.547 4.512-.815 1.038H0v-5.53l4.301-5.33Z' fill='%2316BDCC' fill-rule='evenodd'%3E%3C/path%3E%3C/svg%3E");
  -webkit-mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 29 26'%3E%3Cpath clip-rule='evenodd' d='m9.144 21.802.693-.894h5.126l2.672 3.415-.475 1.159h-6.61l.158-1.159-1.564-2.52Z' fill='%23169FCC' fill-rule='evenodd'%3E%3C/path%3E%3Cpath clip-rule='evenodd' d='m10.711 10.158 2.476-4.278h2.884L29 20.903l-1.003 4.579h-5.543l.347-1.286-12.09-14.038Z' fill='%2314B59A' fill-rule='evenodd'%3E%3C/path%3E%3Cpath clip-rule='evenodd' d='m13.717 5.068 1.185-2.193L18.457 0l6.093 4.778-.79.807 1.06 1.468v1.572l-3.036 2.482-5.094-6.04h-2.973Z' fill='%231BC263' fill-rule='evenodd'%3E%3C/path%3E%3Cpath clip-rule='evenodd' d='M4.301 14.621h1.61l4.2-3.514 5.565 6.469-1.57 2.356H8.97l-3.547 4.512-.815 1.038H0v-5.53l4.301-5.33Z' fill='%2316BDCC' fill-rule='evenodd'%3E%3C/path%3E%3C/svg%3E");mask-size: contain;
  -webkit-mask-size: contain;
  mask-repeat: no-repeat;
  -webkit-mask-repeat: no-repeat;
}

.navbar-controls {
  display: flex;
  align-items: center;
  gap: 2rem;
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
@media (max-width: 768px) {
  .navbar {
    padding: 1rem;
  }
  
  .user-name {
    max-width: 100px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

@media (max-width: 480px) {
  .user-name {
    display: none;
  }
}
</style>