import { useState, type FormEvent } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { authApi } from '@/features/auth/api';
import { useAuth } from '@/contexts/AuthContext';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import styles from './AuthPage.module.css';

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { setUser } = useAuth();
  const from = (location.state as { from?: Location })?.from?.pathname ?? '/dashboard';

  const [form, setForm] = useState({ username: '', password: '' });
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setIsLoading(true);
    try {
      const res = await authApi.login({ username: form.username, password: form.password });
      setUser(res.user);
      navigate(from, { replace: true });
    } catch {
      setError('Usuario o contraseña incorrectos.');
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className={styles.form}>
      <h2 className={styles.title}>Iniciar sesión</h2>

      <Input
        label="Usuario"
        type="text"
        autoComplete="username"
        required
        value={form.username}
        onChange={(e) => setForm((f) => ({ ...f, username: e.target.value }))}
      />
      <Input
        label="Contraseña"
        type="password"
        autoComplete="current-password"
        required
        value={form.password}
        onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
      />

      {error && <p className={styles.error}>{error}</p>}

      <Button type="submit" isLoading={isLoading} className={styles.submit}>
        Entrar
      </Button>

      <p className={styles.link}>
        ¿No tienes cuenta?{' '}
        <Link to="/signup">Regístrate gratis</Link>
      </p>
    </form>
  );
}
