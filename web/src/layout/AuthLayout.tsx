import { Link, Outlet } from 'react-router-dom';
import { LogoMark } from '@/components/LogoMark';
import styles from './AuthLayout.module.css';

export function AuthLayout() {
  return (
    <div className={styles.container}>
      <div className={styles.glow} />
      <div className={styles.card}>
        <Link to="/" className={styles.brand}>
          <LogoMark size={24} />
          <span className={styles.brandName}>Your Own Boss</span>
        </Link>
        <Outlet />
      </div>
    </div>
  );
}
