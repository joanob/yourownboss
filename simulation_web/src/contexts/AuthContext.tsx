import React, {createContext, useContext, useEffect, useState} from 'react'
import api from '../lib/api'

type AuthContextType = {
  token: string | null
  setToken: (t: string | null) => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<React.PropsWithChildren<{}>> = ({children}) => {
  // token is kept in memory only; do not persist across page reloads
  const [token, setTokenState] = useState<string | null>(null)

  useEffect(() => {
    if (token) {
      // Set Authorization only on our api axios instance
      try { api.defaults.headers.common['Authorization'] = `Bearer ${token}` } catch {}
    } else {
      try { delete api.defaults.headers.common['Authorization'] } catch {}
    }
  }, [token])

  const setToken = (t: string | null) => setTokenState(t)

  return (
    <AuthContext.Provider value={{token, setToken}}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

export default AuthContext
