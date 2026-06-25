import { createContext, useContext, useReducer, useEffect, type ReactNode } from 'react';
import { usersApi } from '@/features/users/api';
import type { User } from '@/types/api';

interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}

type AuthAction = { type: 'SET_USER'; payload: User } | { type: 'CLEAR_USER' };

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'SET_USER':
      return { user: action.payload, isAuthenticated: true, isLoading: false };
    case 'CLEAR_USER':
      return { user: null, isAuthenticated: false, isLoading: false };
    default:
      return state;
  }
}

interface AuthContextValue extends AuthState {
  setUser: (user: User) => void;
  clearUser: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(authReducer, {
    user: null,
    isLoading: true,
    isAuthenticated: false,
  });

  useEffect(() => {
    usersApi
      .getMe()
      .then((user) => dispatch({ type: 'SET_USER', payload: user }))
      .catch(() => dispatch({ type: 'CLEAR_USER' }));
  }, []);

  const setUser = (user: User) => dispatch({ type: 'SET_USER', payload: user });
  const clearUser = () => dispatch({ type: 'CLEAR_USER' });

  return (
    <AuthContext.Provider value={{ ...state, setUser, clearUser }}>{children}</AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth debe usarse dentro de AuthProvider');
  return ctx;
}
