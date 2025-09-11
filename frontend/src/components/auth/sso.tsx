import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useKeycloak } from '@react-keycloak/web';
import { Loader2 } from 'lucide-react';

export function SSOHandler() {
  const { keycloak, initialized } = useKeycloak();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  
  useEffect(() => {
    if (!initialized || !keycloak) return;
    
    // Get token and client_id from URL parameters
    const token = searchParams.get('token');
    const clientId = searchParams.get('client_id');
    
    if (token) {
      // If we have a token from Django, use it to authenticate
      // This is for programmatic token passing
      handleTokenAuth(token, clientId);
    } else if (clientId) {
      // Store client ID for context
      localStorage.setItem('client_id', clientId);
      
      // Check if user is already authenticated via Keycloak SSO
      if (keycloak.authenticated) {
        // User is already logged in via Keycloak SSO
        navigate('/dashboard', { replace: true });
      } else {
        // Try silent authentication first
        keycloak.login({ 
          prompt: 'none',
          redirectUri: window.location.origin + '/dashboard'
        }).catch(() => {
          // If silent auth fails, redirect to login
          navigate('/login', { replace: true });
        });
      }
    } else {
      // No special parameters, check normal SSO
      if (keycloak.authenticated) {
        navigate('/dashboard', { replace: true });
      } else {
        navigate('/login', { replace: true });
      }
    }
  }, [initialized, keycloak, searchParams, navigate]);
  
  const handleTokenAuth = async (token: string, clientId: string | null) => {
    try {
      // Store the token in Keycloak's token store
      // This is a simplified approach - in production you might want to
      // validate the token first or exchange it properly
      
      // Set the token on the Keycloak instance
      (keycloak as any).token = token;
      (keycloak as any).authenticated = true;
      
      // Store client ID if provided
      if (clientId) {
        localStorage.setItem('client_id', clientId);
      }
      
      // Navigate to dashboard
      navigate('/dashboard', { replace: true });
    } catch (error) {
      console.error('Failed to handle SSO token:', error);
      navigate('/login', { replace: true });
    }
  };
  
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4 text-blue-600" />
        <p className="text-gray-600">Completing sign in...</p>
      </div>
    </div>
  );
}