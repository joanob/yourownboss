import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/Button';
import styles from './LandingPage.module.css';

export function LandingPage() {
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <div className={styles.logo}>
          <span className={styles.logoIcon}>🏢</span>
          <span className={styles.logoName}>Your Own Boss</span>
        </div>
        <nav className={styles.nav}>
          <Link to="/login">
            <Button variant="ghost" size="sm">Iniciar sesión</Button>
          </Link>
          <Link to="/signup">
            <Button size="sm">Empezar gratis</Button>
          </Link>
        </nav>
      </header>

      <main className={styles.main}>
        <section className={styles.hero}>
          <h1 className={styles.heroTitle}>
            Construye tu empresa.<br />
            <span className={styles.heroAccent}>Sé tu propio jefe.</span>
          </h1>
          <p className={styles.heroSubtitle}>
            Gestiona recursos, produce bienes, domina el mercado y haz crecer tu empresa desde cero en un juego de simulación económica por turnos.
          </p>
          <div className={styles.heroCta}>
            <Link to="/signup">
              <Button size="lg">Crear cuenta gratis</Button>
            </Link>
            <Link to="/login">
              <Button variant="secondary" size="lg">Ya tengo cuenta</Button>
            </Link>
          </div>
        </section>

        <section className={styles.features}>
          <div className={styles.feature}>
            <span className={styles.featureIcon}>🏭</span>
            <h3>Producción</h3>
            <p>Construye y mejora edificios para fabricar productos con tus recursos.</p>
          </div>
          <div className={styles.feature}>
            <span className={styles.featureIcon}>📈</span>
            <h3>Mercado</h3>
            <p>Compra materias primas y vende tus productos al mejor precio.</p>
          </div>
          <div className={styles.feature}>
            <span className={styles.featureIcon}>💼</span>
            <h3>Empresa</h3>
            <p>Gestiona tu compañía, expande operaciones y maximiza beneficios.</p>
          </div>
        </section>
      </main>
    </div>
  );
}
