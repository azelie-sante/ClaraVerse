import { Navigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/store/useAuthStore';
import { Spinner } from '@/components/design-system';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuthStore();
  const location = useLocation();

  if (isLoading) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: '100vh',
        }}
      >
        <Spinner size="lg" />
      </div>
    );
  }

  if (!isAuthenticated) {
    // Capture the current URL (path + search params) for redirect after login
    const currentUrl = location.pathname + location.search;
    const redirectUrl = currentUrl !== '/' ? `?redirect=${encodeURIComponent(currentUrl)}` : '';
    return <Navigate to={`/signin${redirectUrl}`} replace />;
  }

  return <>{children}</>;
};
