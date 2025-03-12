import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';

// Create pinia store
const pinia = createPinia();

// Create app
const app = createApp(App);

// Use plugins
app.use(pinia);
app.use(router);

// Mount app
app.mount('#app');