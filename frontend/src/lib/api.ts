import axios from 'axios';
import type Keycloak from 'keycloak-js';

// Detect if running in Docker (when hostname is not localhost)
const isDocker = window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1';
const BACKEND_URL = isDocker ? 'http://debt-backend:8080' : 'http://localhost:8080';
const API_BASE_URL = `${BACKEND_URL}/api`;

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const setupInterceptors = (keycloak: Keycloak) => {
  console.log('Setting up API interceptors with Keycloak:', !!keycloak);
  
  apiClient.interceptors.request.use(
    (config) => {
      console.log('API Request interceptor - Token available:', !!keycloak.token);
      console.log('API Request interceptor - Authenticated:', !!keycloak.authenticated);
      
      if (keycloak.token) {
        config.headers.Authorization = `Bearer ${keycloak.token}`;
        console.log('Added Authorization header to request');
      } else {
        console.warn('No token available for API request');
      }
      return config;
    },
    (error) => {
      console.error('API Request interceptor error:', error);
      return Promise.reject(error);
    }
  );

  apiClient.interceptors.response.use(
    (response) => response,
    (error) => {
      console.error('API Response interceptor error:', {
        status: error.response?.status,
        statusText: error.response?.statusText,
        data: error.response?.data,
        url: error.config?.url,
      });
      
      if (error.response?.status === 401) {
        console.warn('401 Unauthorized - logging out user');
        // Instead of redirecting, we can just log out.
        // The ProtectedRoute component will handle the redirect to the login page.
        keycloak.logout();
      }
      return Promise.reject(error);
    }
  );
};

// API endpoints
export const api = {
  // Authentication
  auth: {
    config: () => axios.get(`${BACKEND_URL}/auth/config`), // Use dynamic backend URL
  },

  // File uploads
  uploads: {
    create: (file: File) => {
      const formData = new FormData();
      formData.append('file', file);
      return apiClient.post('/uploads', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
    },
    getById: (id: string) => apiClient.get(`/uploads/${id}`),
    list: () => apiClient.get('/uploads'),
  },

  // Accounts
  accounts: {
    getByUploadId: (uploadId: string) => apiClient.get(`/uploads/${uploadId}/accounts`),
  },

  // Messaging
  messaging: {
    send: (accountIds: string[]) => 
      apiClient.post('/messaging/send', { account_ids: accountIds }),
    getLogs: () => apiClient.get('/logs/messaging'),
  },
};

export default apiClient;
