<template>
    <div class="services-panel">
      <div class="header-container">
        <h3>Services</h3>
      </div>
      
      <div v-if="loading" class="loading-container">
        <div class="spinner"></div>
        <p>Loading services...</p>
      </div>
      
      <div v-else-if="error" class="error-message">
        {{ error }}
      </div>
      
      <div v-else-if="services.length === 0" class="empty-message">
        No services found
      </div>
      
      <div v-else class="services-list">
        <div 
          v-for="service in services" 
          :key="service.name"
          class="service-card"
        >
          <h4 class="service-name">{{ service.name }}</h4>
          <p v-if="service.description" class="service-description">{{ service.description }}</p>
          
          <div class="service-meta">
            <div class="service-branch">
              <span class="branch-label">Prod:</span>
              <span class="branch-name">{{ service.prodBranch }}</span>
            </div>
            <div class="service-branch">
              <span class="branch-label">Dev:</span>
              <span class="branch-name">{{ service.devBranch }}</span>
            </div>
            <div v-if="service.specPath" class="service-path">
              <span class="path-label">Spec:</span>
              <span class="path-value">{{ service.specPath }}</span>
            </div>
          </div>
          
          <div v-if="service.git && service.git.repo" class="service-repo">
            <a 
              :href="getRepoUrl(service.git)" 
              target="_blank" 
              rel="noopener noreferrer"
              class="repo-link"
            >
              View Repository
            </a>
          </div>
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import { ref, onMounted } from 'vue';
  import api from '@/services/api';
  
  // State
  const services = ref([]);
  const loading = ref(true);
  const error = ref(null);
  const useDummyData = ref(true); // Set to true to use dummy data
  
  // Dummy data for demonstration
  const dummyServices = [
    {
      name: "user-service",
      description: "Handles user authentication and profile management",
      git: {
        repo: "acme-org/user-service"
      },
      specPath: "/api/v1/spec",
      prodBranch: "main",
      devBranch: "develop"
    },
    {
      name: "payment-processor",
      description: "Processes payments and manages billing information",
      git: {
        repo: "acme-org/payment-processor"
      },
      specPath: "/specs/payment-api",
      prodBranch: "production",
      devBranch: "development"
    },
    {
      name: "notification-service",
      description: "Sends email, SMS, and push notifications to users",
      git: {
        repo: "acme-org/notification-service"
      },
      specPath: "/docs/api",
      prodBranch: "main",
      devBranch: "dev"
    },
    {
      name: "inventory-manager",
      description: "Manages product inventory and stock levels",
      git: {
        repo: "acme-org/inventory-manager"
      },
      specPath: "/api/specs",
      prodBranch: "master",
      devBranch: "development"
    },
    {
      name: "analytics-service",
      description: "Collects and processes user analytics data",
      git: {
        repo: "acme-org/analytics-service"
      },
      specPath: "/api/docs",
      prodBranch: "main",
      devBranch: "dev"
    }
  ];
  
  // Methods
  const fetchServices = async () => {
    loading.value = true;
    error.value = null;
    
    // If using dummy data, return that instead
    if (useDummyData.value) {
      setTimeout(() => {
        services.value = dummyServices;
        loading.value = false;
      }, 800); // Simulate API delay
      return;
    }
    
    try {
      // Use the API to get services
      const response = await api.services.getServices();
      services.value = response.data || [];
    } catch (err) {
      console.error('Error fetching services:', err);
      error.value = err.response?.data?.error || 'Failed to load services';
      
      // Use dummy data as fallback if needed
      if (services.value.length === 0) {
        services.value = dummyServices;
        error.value = null; // Clear error since we're showing dummy data
      }
    } finally {
      loading.value = false;
    }
  };
  
  const getRepoUrl = (git) => {
    if (!git || !git.repo) return '#';
    
    // Handle different git providers
    if (git.repo.startsWith('https://') || git.repo.startsWith('http://')) {
      return git.repo;
    }
    
    // Assume GitHub if no full URL provided
    return `https://github.com/${git.repo}`;
  };
  
  // Fetch services on component mount
  onMounted(() => {
    fetchServices();
  });
  </script>
  
  <style scoped>
  .services-panel {
    width: 100%;
    max-width: 100%;
  }
  
  .header-container {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1rem;
  }
  
  h3 {
    margin: 0;
    font-size: 18px;
    color: #24292e;
  }
  
  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 2rem 0;
  }
  
  .spinner {
    width: 24px;
    height: 24px;
    border: 2px solid rgba(0, 0, 0, 0.1);
    border-radius: 50%;
    border-top-color: #0366d6;
    animation: spin 1s linear infinite;
    margin-bottom: 0.5rem;
  }
  
  .error-message {
    color: #cb2431;
    margin-top: 0.5rem;
    padding: 1rem;
    background-color: #ffdce0;
    border-radius: 6px;
    font-size: 14px;
  }
  
  .empty-message {
    color: #586069;
    margin-top: 0.5rem;
    font-style: italic;
    font-size: 14px;
    padding: 2rem 0;
    text-align: center;
    background-color: #f6f8fa;
    border-radius: 6px;
    border: 1px dashed #e1e4e8;
  }
  
  .services-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    max-height: calc(100vh - 12rem);
    overflow-y: auto;
  }
  
  .service-card {
    background-color: #f6f8fa;
    border: 1px solid #e1e4e8;
    border-radius: 6px;
    padding: 1rem;
    transition: transform 0.2s, box-shadow 0.2s;
  }
  
  .service-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  }
  
  .service-name {
    margin: 0 0 0.75rem 0;
    font-size: 16px;
    color: #24292e;
  }
  
  .service-description {
    margin: 0 0 1rem 0;
    color: #586069;
    font-size: 14px;
    line-height: 1.5;
  }
  
  .service-meta {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1rem;
    font-size: 13px;
  }
  
  .service-branch, .service-path {
    display: flex;
    align-items: center;
  }
  
  .branch-label, .path-label {
    font-weight: 600;
    color: #24292e;
    width: 45px;
  }
  
  .branch-name, .path-value {
    color: #586069;
  }
  
  .service-repo {
    margin-top: 0.75rem;
  }
  
  .repo-link {
    display: inline-block;
    color: #0366d6;
    text-decoration: none;
    font-size: 14px;
  }
  
  .repo-link:hover {
    text-decoration: underline;
  }
  
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
  </style>