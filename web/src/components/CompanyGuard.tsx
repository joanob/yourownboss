import { Navigate, Outlet } from 'react-router-dom';
import { useCompany } from '@/features/company/hooks/useCompany';
import { ApiError } from '@/services/api';
import { PageSpinner } from '@/components/ui/Spinner';

export function CompanyGuard() {
  const { data: company, isLoading, error } = useCompany();

  if (isLoading) return <PageSpinner />;

  if (error instanceof ApiError && error.status === 403) {
    return <Navigate to="/setup" replace />;
  }

  if (!company) return null;

  return <Outlet />;
}
