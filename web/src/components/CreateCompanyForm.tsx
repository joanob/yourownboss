import { useState } from 'react';
import { companyAPI } from '@/lib/api';
import type { Company } from '@/lib/api';

interface CreateCompanyFormProps {
  onCompanyCreated: (company: Company) => void;
}

export function CreateCompanyForm({
  onCompanyCreated,
}: CreateCompanyFormProps) {
  const [name, setName] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (name.length < 3 || name.length > 50) {
      setError('Company name must be between 3 and 50 characters');
      return;
    }

    setLoading(true);

    try {
      const response = await companyAPI.createCompany(name);
      onCompanyCreated(response.data);
    } catch (err: any) {
      setError(
        err.response?.data || 'Failed to create company. Please try again.'
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '20px',
        backgroundColor: '#f5f5f5',
      }}
    >
      <div
        style={{
          width: '100%',
          maxWidth: '400px',
          padding: '24px',
          backgroundColor: '#ffffff',
          borderRadius: '12px',
          boxShadow: '0 2px 10px rgba(0,0,0,0.1)',
        }}
      >
        <div style={{ textAlign: 'center', marginBottom: '24px' }}>
          <span
            style={{ fontSize: '48px', display: 'block', marginBottom: '16px' }}
          >
            🏢
          </span>
          <h2 style={{ margin: '0 0 8px 0', fontSize: '22px', color: '#333' }}>
            Crea tu empresa
          </h2>
          <p style={{ margin: 0, fontSize: '14px', color: '#666' }}>
            Necesitas una empresa para empezar a jugar
          </p>
        </div>

        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: '20px' }}>
            <label
              htmlFor="name"
              style={{
                display: 'block',
                marginBottom: '8px',
                fontWeight: 500,
                fontSize: '14px',
              }}
            >
              Nombre de la empresa
            </label>
            <input
              type="text"
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Ej: Mi Empresa Global"
              disabled={loading}
              style={{
                width: '100%',
                padding: '12px',
                fontSize: '16px',
                border: '1px solid #ddd',
                borderRadius: '6px',
                boxSizing: 'border-box',
              }}
              autoFocus
            />
            <small
              style={{
                color: '#666',
                fontSize: '12px',
                display: 'block',
                marginTop: '4px',
              }}
            >
              Entre 3 y 50 caracteres
            </small>
          </div>

          {error && (
            <div
              style={{
                padding: '12px',
                marginBottom: '16px',
                backgroundColor: '#fee',
                color: '#c33',
                border: '1px solid #fcc',
                borderRadius: '6px',
                fontSize: '14px',
              }}
            >
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            style={{
              width: '100%',
              padding: '14px',
              fontSize: '16px',
              backgroundColor: loading ? '#999' : '#007bff',
              color: 'white',
              border: 'none',
              borderRadius: '6px',
              cursor: loading ? 'not-allowed' : 'pointer',
              fontWeight: 600,
            }}
          >
            {loading ? 'Creando...' : 'Crear empresa'}
          </button>
        </form>
      </div>
    </div>
  );
}
