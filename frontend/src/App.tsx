import { BrowserRouter, Routes, Route, Link, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './hooks/useAuth';
import Login from './pages/Login';
import Register from './pages/Register';
import MakePicks from './pages/MakePicks';
import Leaderboard from './pages/Leaderboard';
import Admin from './pages/Admin';

function Navigation() {
  const { user, logout } = useAuth();

  if (!user) return null;

  return (
    <nav className="bg-white shadow-sm border-b">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex space-x-8">
            <Link
              to="/"
              className="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 border-b-2 border-transparent hover:border-gray-300"
            >
              Make Picks
            </Link>
            <Link
              to="/leaderboard"
              className="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 border-b-2 border-transparent hover:border-gray-300"
            >
              Leaderboard
            </Link>
            {user.is_admin && (
              <Link
                to="/admin"
                className="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 border-b-2 border-transparent hover:border-gray-300"
              >
                Admin
              </Link>
            )}
          </div>
          <div className="flex items-center space-x-4">
            <span className="text-sm text-gray-700">
              {user.display_name || user.username}
            </span>
            <button
              onClick={logout}
              className="text-sm text-gray-700 hover:text-gray-900"
            >
              Logout
            </button>
          </div>
        </div>
      </div>
    </nav>
  );
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth();

  if (loading) {
    return <div className="text-center py-8">Loading...</div>;
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

function AppRoutes() {
  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <MakePicks />
            </ProtectedRoute>
          }
        />
        <Route
          path="/leaderboard"
          element={
            <ProtectedRoute>
              <Leaderboard />
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin"
          element={
            <ProtectedRoute>
              <Admin />
            </ProtectedRoute>
          }
        />
      </Routes>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <AppRoutes />
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
