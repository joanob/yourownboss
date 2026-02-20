import { Link } from 'react-router-dom';

export function LandingPage() {
  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        alignItems: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        color: 'white',
        padding: '20px',
      }}
    >
      <div style={{ textAlign: 'center', maxWidth: '600px' }}>
        <h1 style={{ fontSize: '3rem', marginBottom: '1rem' }}>
          Your Own Boss
        </h1>
        <p style={{ fontSize: '1.25rem', marginBottom: '2rem', opacity: 0.9 }}>
          Un juego idle de gestión y producción de recursos. Crea tu empresa,
          gestiona recursos y construye tu imperio.
        </p>

        <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center' }}>
          <Link
            to="/login"
            style={{
              padding: '12px 32px',
              backgroundColor: 'white',
              color: '#667eea',
              textDecoration: 'none',
              borderRadius: '8px',
              fontWeight: 'bold',
              fontSize: '1.1rem',
              transition: 'transform 0.2s',
            }}
            onMouseOver={(e) =>
              (e.currentTarget.style.transform = 'scale(1.05)')
            }
            onMouseOut={(e) => (e.currentTarget.style.transform = 'scale(1)')}
          >
            Login
          </Link>

          <Link
            to="/register"
            style={{
              padding: '12px 32px',
              backgroundColor: 'transparent',
              color: 'white',
              textDecoration: 'none',
              borderRadius: '8px',
              fontWeight: 'bold',
              fontSize: '1.1rem',
              border: '2px solid white',
              transition: 'transform 0.2s',
            }}
            onMouseOver={(e) =>
              (e.currentTarget.style.transform = 'scale(1.05)')
            }
            onMouseOut={(e) => (e.currentTarget.style.transform = 'scale(1)')}
          >
            Register
          </Link>
        </div>
      </div>
    </div>
  );
}
