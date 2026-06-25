import styles from './Spinner.module.css';

interface SpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function Spinner({ size = 'md', className }: SpinnerProps) {
  return (
    <div
      className={[styles.spinner, styles[size], className ?? ''].filter(Boolean).join(' ')}
      role="status"
      aria-label="Cargando"
    />
  );
}

export function PageSpinner() {
  return (
    <div className={styles.page}>
      <Spinner size="lg" />
    </div>
  );
}
