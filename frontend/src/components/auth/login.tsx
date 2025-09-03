import { useKeycloak } from '@react-keycloak/web';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { LogIn } from 'lucide-react';

export function Login() {
  const { keycloak } = useKeycloak();

  const handleLogin = () => {
    keycloak?.login();
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">
            Age Analysis Messaging Application
          </h1>
          <p className="text-gray-600">
            Streamline your debt collection process with automated messaging
          </p>
        </div>

        <Card>
          <CardHeader className="text-center">
            <CardTitle>Welcome Back</CardTitle>
            <CardDescription>
              Please sign in to access your account
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button 
              onClick={handleLogin}
              className="w-full"
              size="lg"
            >
              <LogIn className="mr-2 h-4 w-4" />
              Sign in with Keycloak
            </Button>
          </CardContent>
        </Card>

        <div className="text-center text-sm text-gray-500">
          Secure authentication powered by Keycloak
        </div>
      </div>
    </div>
  );
}