<template>
  <div class="app">
    <NavBar v-if="shouldShowNavBar" />
    <main class="app-content">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import NavBar from '@/components/NavBar.vue';

const route = useRoute();

// Only show navbar on routes where it makes sense
// Don't show on auth callback pages
const shouldShowNavBar = computed(() => {
  return !route.path.includes('/auth/success');
});
</script>

<style>
/* Global styles */
:root {
  --color-bg-dark: #0d1117;
  --color-bg-medium: #1a1f24;
  --color-bg-light: #24292e;
  --color-primary: #00b969;
  --color-secondary: #0366d6;
  --color-accent-blue: #1e90ff;
  --color-text-white: #ffffff;
  --color-text-light: #f1f1f1;
  --color-text-muted: #8b949e;
  --color-border: #30363d;
  --color-gradient-start: var(--color-accent-blue);
  --color-gradient-end: var(--color-primary);
  
  font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  line-height: 1.5;
  font-weight: 400;

  background-color: var(--color-bg-dark);
  color: var(--color-text-white);

  font-synthesis: none;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

body {
  margin: 0;
  padding: 0;
  background-color: var(--color-bg-dark);
  color: var(--color-text-white);
  min-height: 100vh;
}

h1, h2, h3, h4, h5, h6 {
  color: var(--color-text-white);
  line-height: 1.2;
  margin: 0 0 1rem;
}

a {
  color: var(--color-accent-blue);
  text-decoration: none;
  transition: color 0.2s;
}

a:hover {
  color: var(--color-primary);
}

button {
  cursor: pointer;
  font-family: inherit;
  font-size: 14px;
  font-weight: 500;
  padding: 8px 16px;
  border-radius: 4px;
  transition: all 0.2s;
}

.btn-primary {
  background-color: var(--color-secondary);
  color: var(--color-text-white);
  border: none;
}

.btn-primary:hover {
  background-color: #0250a4;
}

.btn-outline {
  background-color: transparent;
  color: var(--color-text-white);
  border: 1px solid var(--color-text-white);
}

.btn-outline:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.gradient-text {
  background: linear-gradient(90deg, var(--color-gradient-start), var(--color-gradient-end));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  color: transparent;
}

.app {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: var(--color-bg-dark);
}

.app-content {
  flex: 1;
  max-width: 1600px;
  margin: 0 auto;
  width: 100%;
  padding: 2rem;
  box-sizing: border-box;
}

/* Apply responsive design */
@media (max-width: 992px) {
  .app-content {
    padding: 1rem;
  }
}
</style>