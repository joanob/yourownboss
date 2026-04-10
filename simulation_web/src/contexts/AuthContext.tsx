import React, {createContext, useContext, useEffect, useState} from 'react'
import axios from 'axios'

type AuthContextType = {
  token: string | null
  setToken: (t: string | null) => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<React.PropsWithChildren<{}>> = ({children}) => {
  const [token, setTokenState] = useState<string | null>(() => {
    try {
      return sessionStorage.getItem('yb_token')
    } catch {
      return null
    }
  })

  useEffect(() => {
    if (token) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
      try { sessionStorage.setItem('yb_token', token) } catch {}
    } else {
      delete axios.defaults.headers.common['Authorization']
      try { sessionStorage.removeItem('yb_token') } catch {}
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
