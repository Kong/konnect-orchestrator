import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';
import authPlugin from './plugins/auth-plugin';

// Create pinia store
const pinia = createPinia();

// Create app
const app = createApp(App);

// Use plugins
app.use(pinia);
app.use(router);
app.use(authPlugin); // Add the auth plugin here

// Mount app
app.mount('#app');