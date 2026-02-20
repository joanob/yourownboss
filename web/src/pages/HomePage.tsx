import { useAuth } from '@/contexts/AuthContext';
import { useNavigate } from 'react-router-dom';

export function HomePage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

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

      <div
        style={{
          padding: '20px',
          backgroundColor: '#f8f9fa',
          borderRadius: '8px',
        }}
      >
        <h2>Dashboard</h2>
        <p>Welcome to Your Own Boss! This is where your game will start.</p>
        <p>
          Authentication is working correctly. You can now build the game
          features.
        </p>
      </div>
    </div>
  );
}
