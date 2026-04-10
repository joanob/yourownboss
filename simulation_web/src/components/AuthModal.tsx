import React, {useState} from 'react'
import {useAuth} from '../contexts/AuthContext'

const pageStyle: React.CSSProperties = {
  minHeight: '100vh',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  padding: 20,
  background: '#f7f7f8'
}

const boxStyle: React.CSSProperties = {
  background: 'white',
  padding: 28,
  borderRadius: 8,
  width: 420,
  boxShadow: '0 6px 18px rgba(0,0,0,0.06)'
}

export const AuthModal: React.FC = () => {
  const {token, setToken} = useAuth()
  const [value, setValue] = useState('')

  // This component is now a full-page auth screen. It should be rendered
  // exclusively when there is no token (App.tsx enforces that).

  return (
    <div style={pageStyle} role="main">
      <div style={boxStyle}>
        <h2 style={{marginTop: 0}}>Autenticación</h2>
        <p>Introduce la contraseña para acceder a la aplicación.</p>
        <input
          autoFocus
          type="password"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          style={{width: '100%', padding: 10, marginBottom: 12, boxSizing: 'border-box'}}
        />
        <div style={{display: 'flex', justifyContent: 'flex-end', gap: 8}}>
          <button onClick={() => setValue('')}>Limpiar</button>
          <button onClick={() => setToken(value)} style={{padding: '6px 12px'}}>Entrar</button>
        </div>
        <p style={{fontSize: 12, marginTop: 12, color: '#666'}}>La contraseña se mantiene solo en memoria y habrá que introducirla al recargar la página.</p>
      </div>
    </div>
  )
}

export default AuthModal
