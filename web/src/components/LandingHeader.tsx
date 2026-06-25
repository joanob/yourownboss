import { Link } from 'react-router-dom';
import { LogoMark } from './LogoMark';
import styles from './LandingHeader.module.css';

export function LandingHeader() {
  return (
    <header className={styles.header}>
      <h1 className={styles.srOnly}>Your Own Boss</h1>
      <Link to="/" className={styles.logo} aria-label="Your Own Boss — inicio">
        <LogoMark size={24} />
        <span className={styles.logoName} aria-hidden="true">
          Your Own Boss
        </span>
      </Link>
    </header>
  );
}
