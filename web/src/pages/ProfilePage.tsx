import { GameLayout } from '@/components/GameLayout';
import { useAuth } from '@/contexts/AuthContext';
import { useCompany } from '@/contexts/CompanyContext';
import { useNavigate } from 'react-router-dom';

export function ProfilePage() {
  const { user, logout } = useAuth();
  const { company } = useCompany();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  return (
    <GameLayout>
      <div style={{ padding: '16px' }}>
        <div
          style={{
            backgroundColor: '#ffffff',
            borderRadius: '8px',
            padding: '20px',
            marginBottom: '16px',
            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          }}
        >
          <div
            style={{
              width: '64px',
              height: '64px',
              borderRadius: '50%',
              backgroundColor: '#007bff',
              color: 'white',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '32px',
              margin: '0 auto 16px',
            }}
          >
            👤
          </div>
          <h2
            style={{
              margin: '0 0 8px 0',
              fontSize: '20px',
              textAlign: 'center',
            }}
          >
            {user?.username}
          </h2>
          {company && (
            <p
              style={{
                margin: 0,
                fontSize: '14px',
                color: '#666',
                textAlign: 'center',
              }}
            >
              CEO de {company.name}
            </p>
          )}
        </div>

        <div
          style={{
            backgroundColor: '#ffffff',
            borderRadius: '8px',
            overflow: 'hidden',
            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          }}
        >
          <button
            onClick={() => navigate('/settings')}
            style={{
              width: '100%',
              padding: '16px',
              border: 'none',
              borderBottom: '1px solid #f0f0f0',
              backgroundColor: 'transparent',
              textAlign: 'left',
              cursor: 'pointer',
              fontSize: '14px',
              display: 'flex',
              alignItems: 'center',
              gap: '12px',
            }}
          >
            <span style={{ fontSize: '20px' }}>⚙️</span>
            <span>Configuración</span>
          </button>
          <button
            onClick={handleLogout}
            style={{
              width: '100%',
              padding: '16px',
              border: 'none',
              backgroundColor: 'transparent',
              textAlign: 'left',
              cursor: 'pointer',
              fontSize: '14px',
              display: 'flex',
              alignItems: 'center',
              gap: '12px',
              color: '#dc3545',
            }}
          >
            <span style={{ fontSize: '20px' }}>🚪</span>
            <span>Cerrar sesión</span>
          </button>
        </div>
      </div>
    </GameLayout>
  );
}
