import { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import { companyAPI } from '@/lib/api';
import type { Company } from '@/lib/api';
import { CreateCompanyForm } from '@/components/CreateCompanyForm';
import { formatMoney } from '@/lib/money';

export function HomePage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [company, setCompany] = useState<Company | null>(null);
  const [loading, setLoading] = useState(true);
  const [hasCompany, setHasCompany] = useState(false);

  useEffect(() => {
    checkCompany();
  }, []);

  const checkCompany = async () => {
    try {
      const response = await companyAPI.getMyCompany();
      setCompany(response.data);
      setHasCompany(true);
    } catch (err: any) {
      if (err.response?.status === 404) {
        // User doesn't have a company yet
        setHasCompany(false);
      } else {
        console.error('Error checking company:', err);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCompanyCreated = (newCompany: Company) => {
    setCompany(newCompany);
    setHasCompany(true);
  };

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

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
        <div style={{ fontSize: '18px', color: '#666' }}>Loading...</div>
      </div>
    );
  }

  // If user doesn't have a company, force them to create one
  if (!hasCompany) {
    return (
      <div>
        <CreateCompanyForm onCompanyCreated={handleCompanyCreated} />
      </div>
    );
  }

  // User has a company, show the dashboard
  return (
    <div style={{ maxWidth: '800px', margin: '50px auto', padding: '20px' }}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '30px',
        }}
      >
        <h1>Your Own Boss</h1>
        <div>
          <span style={{ marginRight: '15px' }}>
            Welcome, {user?.username}!
          </span>
          <button
            onClick={handleLogout}
            style={{
              padding: '8px 16px',
              backgroundColor: '#dc3545',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Logout
          </button>
        </div>
      </div>

      {company && (
        <div
          style={{
            padding: '20px',
            backgroundColor: '#ffffff',
            borderRadius: '8px',
            marginBottom: '20px',
            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          }}
        >
          <h2 style={{ marginBottom: '15px' }}>{company.name}</h2>
          <div
            style={{
              display: 'flex',
              alignItems: 'baseline',
              gap: '10px',
            }}
          >
            <span style={{ fontSize: '14px', color: '#666' }}>Balance:</span>
            <span
              style={{
                fontSize: '24px',
                fontWeight: 'bold',
                color: '#28a745',
              }}
            >
              ${formatMoney(company.money)}
            </span>
          </div>
        </div>
      )}

      <div
        style={{
          padding: '20px',
          backgroundColor: '#f8f9fa',
          borderRadius: '8px',
        }}
      >
        <h2>Dashboard</h2>
        <p>Welcome to Your Own Boss! This is where your game will start.</p>
        <p>Your company is ready. Game features coming soon!</p>
      </div>
    </div>
  );
}
