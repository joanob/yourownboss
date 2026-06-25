import { useAuth } from '@/contexts/AuthContext';
import styles from './HomePage.module.css';

export function HomePage() {
  const { user } = useAuth();

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <h1 className={styles.title}>Dashboard</h1>
        <p className={styles.subtitle}>Bienvenido de nuevo, {user?.username}</p>
      </div>

      <div className={styles.grid}>
        <div className={styles.card}>
          <span className={styles.cardIcon}>🏭</span>
          <h3>Producción</h3>
          <p>Gestiona tus edificios y procesos de fabricación.</p>
        </div>
        <div className={styles.card}>
          <span className={styles.cardIcon}>📈</span>
          <h3>Mercado</h3>
          <p>Compra recursos y vende tus productos.</p>
        </div>
        <div className={styles.card}>
          <span className={styles.cardIcon}>🏪</span>
          <h3>Ventas</h3>
          <p>Administra tus puntos de venta.</p>
        </div>
      </div>
    </div>
  );
}
