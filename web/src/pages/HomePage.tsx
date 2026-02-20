import { useCompany } from '@/contexts/CompanyContext';
import { CreateCompanyForm } from '@/components/CreateCompanyForm';
import { GameLayout } from '@/components/GameLayout';

export function HomePage() {
  const { company, loading, hasCompany, setCompany } = useCompany();

  const handleCompanyCreated = (newCompany: any) => {
    setCompany(newCompany);
  };

  if (loading) {
    return (
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '100vh',
          backgroundColor: '#f5f5f5',
        }}
      >
        <div style={{ fontSize: '18px', color: '#666' }}>Cargando...</div>
      </div>
    );
  }

  // If user doesn't have a company, force them to create one
  if (!hasCompany) {
    return <CreateCompanyForm onCompanyCreated={handleCompanyCreated} />;
  }

  // User has a company, show the game
  return (
    <GameLayout>
      <div style={{ padding: '16px' }}>
        {/* Welcome card */}
        <div
          style={{
            backgroundColor: '#ffffff',
            borderRadius: '8px',
            padding: '16px',
            marginBottom: '16px',
            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          }}
        >
          <h2 style={{ margin: '0 0 8px 0', fontSize: '18px', color: '#333' }}>
            Bienvenido 👋
          </h2>
          <p
            style={{
              margin: 0,
              fontSize: '14px',
              color: '#666',
              lineHeight: '1.5',
            }}
          >
            Tu empresa está lista para empezar a producir y gestionar recursos.
          </p>
        </div>

        {/* Quick stats */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gap: '12px',
            marginBottom: '16px',
          }}
        >
          <div
            style={{
              backgroundColor: '#ffffff',
              borderRadius: '8px',
              padding: '16px',
              boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
            }}
          >
            <div
              style={{ fontSize: '12px', color: '#666', marginBottom: '4px' }}
            >
              Producción
            </div>
            <div
              style={{ fontSize: '20px', fontWeight: 'bold', color: '#333' }}
            >
              0
            </div>
          </div>
          <div
            style={{
              backgroundColor: '#ffffff',
              borderRadius: '8px',
              padding: '16px',
              boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
            }}
          >
            <div
              style={{ fontSize: '12px', color: '#666', marginBottom: '4px' }}
            >
              Recursos
            </div>
            <div
              style={{ fontSize: '20px', fontWeight: 'bold', color: '#333' }}
            >
              0
            </div>
          </div>
        </div>

        {/* Coming soon */}
        <div
          style={{
            backgroundColor: '#fff9e6',
            border: '1px solid #ffd700',
            borderRadius: '8px',
            padding: '16px',
            textAlign: 'center',
          }}
        >
          <span
            style={{ fontSize: '32px', display: 'block', marginBottom: '8px' }}
          >
            🚧
          </span>
          <div style={{ fontSize: '14px', color: '#856404', fontWeight: 500 }}>
            Próximamente: Sistema de producción y recursos
          </div>
        </div>
      </div>
    </GameLayout>
  );
}
