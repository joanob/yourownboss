import type { ReactNode } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useCompany } from '@/contexts/CompanyContext';
import { formatMoney } from '@/lib/money';

interface GameLayoutProps {
  children: ReactNode;
}

export function GameLayout({ children }: GameLayoutProps) {
  const { company } = useCompany();
  const navigate = useNavigate();
  const location = useLocation();

  const isActive = (path: string) => location.pathname === path;

  const navItems = [
    { path: '/home', label: 'Inicio', icon: '🏠' },
    { path: '/production', label: 'Producción', icon: '⚙️' },
    { path: '/inventory', label: 'Inventario', icon: '📦' },
    { path: '/encyclopedia', label: 'Enciclopedia', icon: '📚' },
  ];

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        minHeight: '100vh',
        backgroundColor: '#f5f5f5',
      }}
    >
      {/* Header */}
      {company && (
        <header
          style={{
            position: 'sticky',
            top: 0,
            zIndex: 10,
            backgroundColor: '#ffffff',
            borderBottom: '1px solid #e0e0e0',
            padding: '12px 16px',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            boxShadow: '0 2px 4px rgba(0,0,0,0.05)',
          }}
        >
          {/* Left: Company name + profile icon */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <button
              onClick={() => navigate('/profile')}
              style={{
                width: '36px',
                height: '36px',
                borderRadius: '50%',
                backgroundColor: '#007bff',
                color: 'white',
                border: 'none',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '18px',
                cursor: 'pointer',
                flexShrink: 0,
              }}
              aria-label="Profile"
            >
              👤
            </button>
            <div style={{ overflow: 'hidden' }}>
              <h1
                style={{
                  margin: 0,
                  fontSize: '16px',
                  fontWeight: 600,
                  color: '#333',
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                }}
              >
                {company.name}
              </h1>
            </div>
          </div>

          {/* Right: Money */}
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'flex-end',
              minWidth: '100px',
            }}
          >
            <span
              style={{ fontSize: '10px', color: '#666', marginBottom: '2px' }}
            >
              Balance
            </span>
            <span
              style={{
                fontSize: '14px',
                fontWeight: 'bold',
                color: '#28a745',
                whiteSpace: 'nowrap',
              }}
            >
              ${formatMoney(company.money)}
            </span>
          </div>
        </header>
      )}

      {/* Main content */}
      <main
        style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          minHeight: 0,
          overflowY: 'auto',
        }}
      >
        {children}
      </main>

      {/* Bottom Navigation */}
      <nav
        style={{
          position: 'fixed',
          bottom: 0,
          left: 0,
          right: 0,
          backgroundColor: '#ffffff',
          borderTop: '1px solid #e0e0e0',
          display: 'flex',
          justifyContent: 'space-around',
          padding: '8px 0',
          boxShadow: '0 -2px 4px rgba(0,0,0,0.05)',
        }}
      >
        {navItems.map((item) => (
          <button
            key={item.path}
            onClick={() => navigate(item.path)}
            style={{
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              gap: '4px',
              padding: '8px',
              border: 'none',
              backgroundColor: 'transparent',
              cursor: 'pointer',
              color: isActive(item.path) ? '#007bff' : '#666',
              transition: 'color 0.2s',
            }}
          >
            <span style={{ fontSize: '24px' }}>{item.icon}</span>
            <span
              style={{
                fontSize: '11px',
                fontWeight: isActive(item.path) ? 600 : 400,
              }}
            >
              {item.label}
            </span>
          </button>
        ))}
      </nav>
    </div>
  );
}
