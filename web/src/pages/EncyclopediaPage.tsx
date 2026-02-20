import { GameLayout } from '@/components/GameLayout';

export function EncyclopediaPage() {
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
            📚
          </span>
          <h2 style={{ margin: '0 0 8px 0', fontSize: '20px', color: '#333' }}>
            Enciclopedia
          </h2>
          <p style={{ margin: 0, fontSize: '14px', color: '#666' }}>
            Consulta información sobre recursos, edificios y procesos de
            producción.
          </p>
        </div>
      </div>
    </GameLayout>
  );
}
