import { NavLink, useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { useAuth } from '@/contexts/AuthContext';
import { useCompany } from '@/features/company/hooks/useCompany';
import { authApi } from '@/features/auth/api';
import { formatMoney } from '@/lib/formatters';
import styles from './Sidebar.module.css';

const navItems = [
  { to: '/dashboard', label: 'Panel', icon: '📊' },
  { to: '/market', label: 'Mercado', icon: '🛒' },
  { to: '/production', label: 'Producción', icon: '🏭' },
  { to: '/sale', label: 'Ventas', icon: '💰' },
];

export function Sidebar() {
  const { user, clearUser } = useAuth();
  const { data: company } = useCompany();
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await authApi.logout();
    } catch {
      // ignorar errores de logout
    }
    clearUser();
    queryClient.clear();
    navigate('/login', { replace: true });
  };

  return (
    <aside className={styles.sidebar}>
      <div className={styles.brand}>
        <span className={styles.brandIcon}>🏢</span>
        <span className={styles.brandName}>Your Own Boss</span>
      </div>

      {company && (
        <div className={styles.companyInfo}>
          <div className={styles.companyName}>{company.name}</div>
          <div className={styles.money}>{formatMoney(company.money)}</div>
        </div>
      )}

      <nav className={styles.nav}>
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              [styles.navItem, isActive ? styles.navItemActive : ''].filter(Boolean).join(' ')
            }
          >
            <span className={styles.navIcon}>{item.icon}</span>
            <span>{item.label}</span>
          </NavLink>
        ))}
      </nav>

      <div className={styles.footer}>
        <NavLink
          to="/profile"
          className={({ isActive }) =>
            [styles.navItem, isActive ? styles.navItemActive : ''].filter(Boolean).join(' ')
          }
        >
          <span className={styles.navIcon}>👤</span>
          <span>Perfil</span>
        </NavLink>
        <div className={styles.userInfo}>
          <span className={styles.username}>{user?.username}</span>
        </div>
        <button className={styles.logoutBtn} onClick={handleLogout}>
          <span className={styles.navIcon}>🚪</span>
          Cerrar sesión
        </button>
      </div>
    </aside>
  );
}
