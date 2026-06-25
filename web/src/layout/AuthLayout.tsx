import { Outlet } from 'react-router-dom';
import styles from './AuthLayout.module.css';

export function AuthLayout() {
  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <div className={styles.brand}>
          <span className={styles.brandIcon}>🏢</span>
          <h1 className={styles.brandName}>Your Own Boss</h1>
          <p className={styles.brandTagline}>Tu empresa, tus reglas</p>
        </div>
        <Outlet />
      </div>
    </div>
  );
}
