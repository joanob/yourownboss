import { Link } from 'react-router-dom';
import { LandingHeader } from '@/components/LandingHeader';
import styles from './LandingPage.module.css';

export function LandingPage() {
  return (
    <div className={styles.page}>
      <LandingHeader />

      <main className={styles.hero}>
        <div className={styles.glow} />
        <h1 className={styles.title}>
          Dirige tu empresa.
          <br />
          <span className={styles.accent}>Sé tu propio jefe.</span>
        </h1>
        <p className={styles.subtitle}>
          Un juego de simulación empresarial por turnos. Produce bienes, compra y vende en el
          mercado y haz crecer tu compañía desde cero.
        </p>
        <div className={styles.cta}>
          <Link to="/signup" className={styles.ctaPrimary}>
            Crear cuenta gratis
          </Link>
          <Link to="/login" className={styles.ctaSecondary}>
            Ya tengo cuenta
          </Link>
        </div>
      </main>
    </div>
  );
}
