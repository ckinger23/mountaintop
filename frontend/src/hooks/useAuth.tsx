import { useState, useEffect, createContext, useContext, ReactNode } from 'react';
import { User } from '../types';
import { authService } from '../services/api';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string, displayName: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadUser = async () => {
      if (authService.isAuthenticated()) {
        try {
          const currentUser = await authService.getCurrentUser();
          setUser(currentUser);
        } catch (error) {
          console.error('Failed to load user:', error);
          authService.logout();
        }
      }
      setLoading(false);
    };

    loadUser();
  }, []);

  const login = async (email: string, password: string) => {
    const response = await authService.login(email, password);
    setUser(response.user);
  };

  const register = async (username: string, email: string, password: string, displayName: string) => {
    const response = await authService.register(username, email, password, displayName);
    setUser(response.user);
  };

  const logout = () => {
    authService.logout();
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
