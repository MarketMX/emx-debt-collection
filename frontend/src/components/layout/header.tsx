import { useKeycloak } from '@react-keycloak/web';
import { useLocation, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { LogOut, User, Home, MessageSquare, Users, Activity } from 'lucide-react';

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
    <header className="bg-white/80 backdrop-blur-lg border-b border-slate-200/60 shadow-sm sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-18 py-3">
          <div className="flex items-center space-x-8">
            {/* Logo/Brand */}
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-cyan-600 rounded-xl flex items-center justify-center">
                <Activity className="h-6 w-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-slate-800">
                  CollectionHub
                </h1>
                <p className="text-xs text-slate-500 leading-none">Medical Debt Collection</p>
              </div>
            </div>
            
            {/* Navigation */}
            <nav className="hidden md:flex space-x-1">
              {navigationItems.map(({ path, label, icon: Icon }) => (
                <button
                  key={path}
                  onClick={() => navigate(path)}
                  className={`flex items-center space-x-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all duration-200 ${
                    location.pathname === path
                      ? 'bg-gradient-to-r from-blue-500 to-cyan-600 text-white shadow-md'
                      : 'text-slate-600 hover:text-slate-800 hover:bg-slate-100'
                  }`}
                >
                  <Icon className="h-4 w-4" />
                  <span>{label}</span>
                </button>
              ))}
            </nav>
          </div>
          
          <div className="flex items-center space-x-4">
            {/* System Status */}
            <div className="hidden sm:flex items-center space-x-2">
              <Badge className="bg-emerald-100 text-emerald-700 hover:bg-emerald-100">
                <div className="w-2 h-2 bg-emerald-500 rounded-full mr-2"></div>
                System Online
              </Badge>
            </div>

            {/* User Info */}
            <div className="flex items-center space-x-3">
              <div className="hidden sm:block text-right">
                <div className="text-sm font-medium text-slate-800">
                  {keycloak?.tokenParsed?.preferred_username || 'User'}
                </div>
                <div className="text-xs text-slate-500">
                  {keycloak?.tokenParsed?.email}
                </div>
              </div>
              <div className="w-8 h-8 bg-gradient-to-br from-slate-200 to-slate-300 rounded-lg flex items-center justify-center">
                <User className="h-4 w-4 text-slate-600" />
              </div>
            </div>
            
            {/* Logout Button */}
            <Button
              variant="outline"
              size="sm"
              onClick={handleLogout}
              className="border-slate-200 hover:bg-slate-50 text-slate-600 hover:text-slate-800"
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