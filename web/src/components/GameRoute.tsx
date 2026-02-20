import type { ReactNode } from 'react';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { CompanyProvider } from '@/contexts/CompanyContext';

interface GameRouteProps {
  children: ReactNode;
}

export function GameRoute({ children }: GameRouteProps) {
  return (
    <ProtectedRoute>
      <CompanyProvider>{children}</CompanyProvider>
    </ProtectedRoute>
  );
}
