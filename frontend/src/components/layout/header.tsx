import { useKeycloak } from '@react-keycloak/web';
import { useLocation, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { LogOut, User, Home, MessageSquare, Users } from 'lucide-react';

export function Header() {
  const { keycloak } = useKeycloak();
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = () => {
    keycloak?.logout();
  };

  const isAdmin = keycloak?.tokenParsed?.realm_access?.roles?.includes('admin') || true; // Demo: always show admin menu

  const navigationItems = [
    { path: '/dashboard', label: 'Dashboard', icon: Home },
    { path: '/templates', label: 'Templates', icon: MessageSquare },
    ...(isAdmin ? [{ path: '/admin/users', label: 'Users', icon: Users }] : []),
  ];

  return (
    <header className="bg-white border-b border-gray-200 shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center space-x-8">
            <h1 className="text-xl font-semibold text-gray-900">
              AAMA
            </h1>
            
            <nav className="hidden md:flex space-x-6">
              {navigationItems.map(({ path, label, icon: Icon }) => (
                <button
                  key={path}
                  onClick={() => navigate(path)}
                  className={`flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                    location.pathname === path
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                  }`}
                >
                  <Icon className="h-4 w-4" />
                  <span>{label}</span>
                </button>
              ))}
            </nav>
          </div>
          
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2 text-sm text-gray-600">
              <User className="h-4 w-4" />
              <span>{keycloak?.tokenParsed?.preferred_username || keycloak?.tokenParsed?.email}</span>
            </div>
            
            <Button
              variant="outline"
              size="sm"
              onClick={handleLogout}
            >
              <LogOut className="h-4 w-4 mr-2" />
              Logout
            </Button>
          </div>
        </div>
      </div>
    </header>
  );
}