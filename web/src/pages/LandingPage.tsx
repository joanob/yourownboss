import { Link } from 'react-router-dom';
import styles from './LandingPage.module.css';

export function LandingPage() {
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <div className={styles.logo}>
          <div className={styles.logoMark}>
            <svg viewBox="0 0 28 28" fill="none">
              <rect width="28" height="28" rx="7" fill="#f06449" />
              <rect x="6" y="6" width="7" height="7" rx="1.5" fill="white" />
              <rect x="15" y="6" width="7" height="7" rx="1.5" fill="white" fillOpacity=".6" />
              <rect x="6" y="15" width="7" height="7" rx="1.5" fill="white" fillOpacity=".6" />
              <rect x="15" y="15" width="7" height="7" rx="1.5" fill="white" fillOpacity=".3" />
            </svg>
          </div>
          <span className={styles.logoName}>Your Own Boss</span>
        </div>
        <nav className={styles.nav}>
          <Link to="/login" className={styles.navLink}>Iniciar sesión</Link>
          <Link to="/signup" className={styles.navCta}>Empezar gratis</Link>
        </nav>
      </header>

      <main className={styles.hero}>
        <div className={styles.glow} />
        <h1 className={styles.title}>
          Dirige tu empresa.<br />
          <span className={styles.accent}>Sé tu propio jefe.</span>
        </h1>
        <p className={styles.subtitle}>
          Un juego de simulación empresarial por turnos. Produce bienes, compra y vende en el mercado y haz crecer tu compañía desde cero.
        </p>
        <div className={styles.cta}>
          <Link to="/signup" className={styles.ctaPrimary}>Crear cuenta gratis</Link>
          <Link to="/login" className={styles.ctaSecondary}>Ya tengo cuenta</Link>
        </div>
      </main>
    </div>
  );
}
