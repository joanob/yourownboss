import React, {useState} from 'react'
import {useAuth} from '../contexts/AuthContext'

const modalStyle: React.CSSProperties = {
  position: 'fixed',
  inset: 0,
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  background: 'rgba(0,0,0,0.4)'
}

const boxStyle: React.CSSProperties = {
  background: 'white',
  padding: 20,
  borderRadius: 8,
  minWidth: 320
}

export const AuthModal: React.FC = () => {
  const {token, setToken} = useAuth()
  const [value, setValue] = useState('')

  if (token) return null

  return (
    <div style={modalStyle} role="dialog" aria-modal>
      <div style={boxStyle}>
        <h3>Introduce la contraseña</h3>
        <input
          autoFocus
          type="password"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          style={{width: '100%', padding: 8, marginBottom: 10}}
        />
        <div style={{display: 'flex', justifyContent: 'flex-end', gap: 8}}>
          <button onClick={() => setToken(null)}>Salir</button>
          <button onClick={() => setToken(value)}>Aceptar</button>
        </div>
        <p style={{fontSize: 12, marginTop: 10}}>
          La contraseña se guarda solo durante la sesión del navegador.
        </p>
      </div>
    </div>
  )
}

export default AuthModal
