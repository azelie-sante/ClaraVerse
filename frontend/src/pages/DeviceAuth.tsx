import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '@/store/useAuthStore';
import { MCPOnboarding } from '@/components/mcp-setup';
import { Card, Spinner } from '@/components/design-system';
import './DeviceAuth.css';

export const DeviceAuth = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { isAuthenticated, isLoading: authLoading } = useAuthStore();

  // Get code from URL if provided
  const codeParam = searchParams.get('code') || undefined;

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      // Preserve the code in the redirect URL
      const redirectUrl = codeParam ? `/device?code=${codeParam}` : '/device';
      navigate(`/signin?redirect=${encodeURIComponent(redirectUrl)}`);
    }
  }, [isAuthenticated, authLoading, navigate, codeParam]);

  const handleComplete = () => {
    navigate('/');
  };

  // Show loading while checking auth
  if (authLoading) {
    return (
      <div className="device-auth-page">
        <Card variant="glass" className="device-auth-loading-card" hoverable={false}>
          <Spinner size="lg" label="Loading..." />
        </Card>
      </div>
    );
  }

  // Only render MCPOnboarding if authenticated
  if (!isAuthenticated) {
    return null;
  }

  return <MCPOnboarding initialCode={codeParam} onComplete={handleComplete} />;
};
