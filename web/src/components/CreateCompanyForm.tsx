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
        maxWidth: '500px',
        margin: '100px auto',
        padding: '30px',
        backgroundColor: '#ffffff',
        borderRadius: '8px',
        boxShadow: '0 2px 10px rgba(0,0,0,0.1)',
      }}
    >
      <h2 style={{ marginBottom: '10px', textAlign: 'center' }}>
        Create Your Company
      </h2>
      <p
        style={{
          textAlign: 'center',
          color: '#666',
          marginBottom: '30px',
        }}
      >
        You need to create a company before you can start playing
      </p>

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: '20px' }}>
          <label
            htmlFor="name"
            style={{
              display: 'block',
              marginBottom: '8px',
              fontWeight: 500,
            }}
          >
            Company Name
          </label>
          <input
            type="text"
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Enter your company name"
            disabled={loading}
            style={{
              width: '100%',
              padding: '10px',
              fontSize: '16px',
              border: '1px solid #ddd',
              borderRadius: '4px',
              boxSizing: 'border-box',
            }}
            autoFocus
          />
          <small style={{ color: '#666', fontSize: '12px' }}>
            3-50 characters
          </small>
        </div>

        {error && (
          <div
            style={{
              padding: '10px',
              marginBottom: '15px',
              backgroundColor: '#f8d7da',
              color: '#721c24',
              border: '1px solid #f5c6cb',
              borderRadius: '4px',
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
            padding: '12px',
            fontSize: '16px',
            backgroundColor: loading ? '#6c757d' : '#007bff',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: loading ? 'not-allowed' : 'pointer',
            fontWeight: 500,
          }}
        >
          {loading ? 'Creating...' : 'Create Company'}
        </button>
      </form>
    </div>
  );
}
