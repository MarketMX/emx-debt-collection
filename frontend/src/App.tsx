import React, { useEffect, useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ReactKeycloakProvider, useKeycloak } from '@react-keycloak/web';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import Keycloak from 'keycloak-js';
import { Dashboard } from '@/components/dashboard/dashboard';
import { Login } from '@/components/auth/login';
import { SSOHandler } from '@/components/auth/sso';
import { MessageTemplates } from '@/components/messaging/message-templates';
import { UserManagement } from '@/components/admin/user-management';
import { MessagingProgress } from '@/components/messaging/messaging-progress';
import { Toaster } from '@/components/ui/toaster';
import { api, setupInterceptors } from '@/lib/api';
import type { AuthConfig } from '@/types';
import { Loader2 } from 'lucide-react';

// Create a query client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

const KeycloakInterceptor = ({ children }: { children: React.ReactNode }) => {
  const { keycloak, initialized } = useKeycloak();

  useEffect(() => {
    if (initialized) {
      setupInterceptors(keycloak);
    }
  }, [initialized, keycloak]);

  return <>{children}</>;
};

function App() {
  const [keycloak, setKeycloak] = useState<Keycloak | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Fetch Keycloak configuration from backend
    api.auth.config()
      .then(response => {
        const config: AuthConfig = response.data;
        
        const keycloakInstance = new Keycloak({
          url: config.auth_url,
          realm: config.realm,
          clientId: config.client_id,
        });
        
        setKeycloak(keycloakInstance);
      })
      .catch(error => {
        console.error('Failed to fetch auth config:', error);
        setError('Failed to load authentication configuration');
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4 text-blue-600" />
          <p className="text-gray-600">Loading application...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="bg-red-50 border border-red-200 rounded-lg p-6 max-w-md mx-auto">
            <h2 className="text-lg font-semibold text-red-800 mb-2">Configuration Error</h2>
            <p className="text-red-600">{error}</p>
            <button 
              onClick={() => window.location.reload()}
              className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition-colors"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (!keycloak) {
    return null;
  }

  return (
    <ReactKeycloakProvider
      authClient={keycloak}
      initOptions={{
        onLoad: 'check-sso',
        checkLoginIframe: false,
      }}
      LoadingComponent={
        <div className="min-h-screen flex items-center justify-center bg-gray-50">
          <div className="text-center">
            <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4 text-blue-600" />
            <p className="text-gray-600">Authenticating...</p>
          </div>
        </div>
      }
    >
      <QueryClientProvider client={queryClient}>
        <Router>
          <KeycloakInterceptor>
            <AppRoutes />
          </KeycloakInterceptor>
        </Router>
        <Toaster />
      </QueryClientProvider>
    </ReactKeycloakProvider>
  );
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/auth/sso" element={<SSOHandler />} />
      <Route path="/login" element={<Login />} />
      <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
      <Route path="/templates" element={<ProtectedRoute><MessageTemplates /></ProtectedRoute>} />
      <Route path="/admin/users" element={<ProtectedRoute><UserManagement /></ProtectedRoute>} />
      <Route path="/messaging/progress" element={<ProtectedRoute><MessagingProgress /></ProtectedRoute>} />
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { keycloak, initialized } = useKeycloak();
  const [isChecking, setIsChecking] = useState(true);

  useEffect(() => {
    if (initialized && keycloak) {
      // Check if token is expired and try to refresh
      if (keycloak.isTokenExpired?.()) {
        keycloak.updateToken(30)
          .then((refreshed) => {
            if (refreshed) {
              console.log('Token refreshed');
            }
            setIsChecking(false);
          })
          .catch(() => {
            console.log('Failed to refresh token');
            keycloak.logout();
          });
      } else {
        setIsChecking(false);
      }
    }
  }, [initialized, keycloak]);

  if (!initialized || isChecking) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4 text-blue-600" />
          <p className="text-gray-600">Checking authentication...</p>
        </div>
      </div>
    );
  }

  if (!keycloak?.authenticated) {
    console.log('User not authenticated, redirecting to login');
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

export default App;
