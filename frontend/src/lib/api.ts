import axios from 'axios';

const API_BASE_URL = '/api/v1';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor to include auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Add response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Clear token and redirect to login
      localStorage.removeItem('access_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// API endpoints
export const api = {
  // Authentication
  auth: {
    config: () => apiClient.get('/auth/config'),
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
    getByUploadId: (uploadId: string) => apiClient.get(`/accounts?upload_id=${uploadId}`),
  },

  // Messaging
  messaging: {
    send: (accountIds: string[]) => 
      apiClient.post('/messaging/send', { account_ids: accountIds }),
    getLogs: () => apiClient.get('/logs/messaging'),
  },
};

export default apiClient;