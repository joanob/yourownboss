import { GameLayout } from '@/components/GameLayout';

export function ProductionPage() {
  return (
    <GameLayout>
      <div style={{ padding: '16px' }}>
        <div
          style={{
            backgroundColor: '#ffffff',
            borderRadius: '8px',
            padding: '24px',
            textAlign: 'center',
            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          }}
        >
          <span
            style={{ fontSize: '48px', display: 'block', marginBottom: '16px' }}
          >
            ⚙️
          </span>
          <h2 style={{ margin: '0 0 8px 0', fontSize: '20px', color: '#333' }}>
            Producción
          </h2>
          <p style={{ margin: 0, fontSize: '14px', color: '#666' }}>
            Aquí podrás gestionar tus procesos de producción y edificios.
          </p>
        </div>
      </div>
    </GameLayout>
  );
}
