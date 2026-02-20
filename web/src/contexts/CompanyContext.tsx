import { createContext, useContext, useState, useEffect } from 'react';
import type { ReactNode } from 'react';
import { companyAPI } from '@/lib/api';
import type { Company } from '@/lib/api';

interface CompanyContextType {
  company: Company | null;
  loading: boolean;
  hasCompany: boolean;
  refreshCompany: () => Promise<void>;
  setCompany: (company: Company) => void;
}

const CompanyContext = createContext<CompanyContextType | undefined>(undefined);

export function CompanyProvider({ children }: { children: ReactNode }) {
  const [company, setCompany] = useState<Company | null>(null);
  const [loading, setLoading] = useState(true);
  const [hasCompany, setHasCompany] = useState(false);

  const refreshCompany = async () => {
    setLoading(true);
    try {
      const response = await companyAPI.getMyCompany();
      setCompany(response.data);
      setHasCompany(true);
    } catch (err: any) {
      if (err.response?.status === 404) {
        setHasCompany(false);
        setCompany(null);
      } else {
        console.error('Error fetching company:', err);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    refreshCompany();
  }, []);

  return (
    <CompanyContext.Provider
      value={{ company, loading, hasCompany, refreshCompany, setCompany }}
    >
      {children}
    </CompanyContext.Provider>
  );
}

export function useCompany() {
  const context = useContext(CompanyContext);
  if (context === undefined) {
    throw new Error('useCompany must be used within CompanyProvider');
  }
  return context;
}
