import { useState, useEffect, createContext, useContext, ReactNode } from 'react';
import { User } from '../types';
import { authService } from '../services/api';


// Creates a React Context for managing user auth across the entire app
// "Authentication Control Center" that any component can tap into

// interface defines what data/functions available to component using auth system
interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string, displayName: string) => Promise<void>;
  logout: () => void;
}
// "container" that will hold and distribute auth data to app
// createContext - create "channel"/"container" for broadcasting data
// This channel will broadcast AuthContextType data or undefined
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// The provider component
export const AuthProvider = ({ children }: { children: ReactNode }) => {
  // State management - the user and whether auth status being checked
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  // check if there's a saved Jwt token and fetches user's data
  useEffect(() => {
    const loadUser = async () => {
      // check local storage for jwt
      if (authService.isAuthenticated()) {
        try {
          // api call to get user info - this validates the token against the current database
          const currentUser = await authService.getCurrentUser();
          setUser(currentUser);
        } catch (error) {
          // clears everything if expired/invalid token
          // This handles cases where:
          // 1. Token is expired
          // 2. Token is invalid
          // 3. Database was reset and user no longer exists
          // 4. User permissions changed (e.g., admin status)
          console.error('Failed to load user, clearing auth:', error);
          authService.logout();
        }
      }
      setLoading(false);
    };

    // call the above function
    loadUser();
  }, []);

  // calls api to login and retrieve jwt token
  const login = async (email: string, password: string) => {
    const response = await authService.login(email, password);
    setUser(response.user);
  };

  // Creates a new account and returns jwt
  const register = async (username: string, email: string, password: string, displayName: string) => {
    const response = await authService.register(username, email, password, displayName);
    setUser(response.user);
  };

  // clears jwt from local storage and clears user from state
  const logout = () => {
    authService.logout();
    setUser(null);
  };

  // wraps your app and makes user, loading, and functions available to child components
  // broadcasts the value to all components inside
  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

// convenience function components use to access authentication
// returns the context value: user, login, logout
export const useAuth = () => {
  // Context is like a "teleportation system" for data
  // Put data at top of your component tree
  // Access data from ANY component without passing props
  // useContext listens to the broadcast
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

// App Loads
// <AuthProvider> runs useEffect
// checks local storage for JWT
// If found, call getCurrentUser()
// set the user state
// Provider roadcasts new value
// All components using useAuth() automatically re-render with new user data
