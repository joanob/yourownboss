import { Navigate } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';

export function PublicRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '100vh',
        }}
      >
        <div>Loading...</div>
      </div>
    );
  }

  // If user is already logged in, redirect to home
  if (user) {
    return <Navigate to="/home" replace />;
  }

  return <>{children}</>;
}
