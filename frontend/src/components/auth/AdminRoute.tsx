import { Navigate } from 'react-router-dom';
import { useAuthStore } from '@/store/useAuthStore';
import { Spinner } from '@/components/design-system';

interface AdminRouteProps {
  children: React.ReactNode;
}

export const AdminRoute: React.FC<AdminRouteProps> = ({ children }) => {
  const { isAuthenticated, isAdmin, isLoading } = useAuthStore();

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
    return <Navigate to="/signin?redirect=/admin" replace />;
  }

  if (!isAdmin) {
    // Non-admin users trying to access admin routes are redirected to home
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
};
