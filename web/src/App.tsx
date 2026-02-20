import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from '@/contexts/AuthContext';
import { GameRoute } from '@/components/GameRoute';
import { PublicRoute } from '@/components/PublicRoute';
import { LandingPage } from '@/pages/LandingPage';
import { LoginPage } from '@/pages/LoginPage';
import { RegisterPage } from '@/pages/RegisterPage';
import { HomePage } from '@/pages/HomePage';
import { ProductionPage } from '@/pages/ProductionPage';
import { InventoryPage } from '@/pages/InventoryPage';
import { EncyclopediaPage } from '@/pages/EncyclopediaPage';
import { ProfilePage } from '@/pages/ProfilePage';

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route
            path="/"
            element={
              <PublicRoute>
                <LandingPage />
              </PublicRoute>
            }
          />
          <Route
            path="/login"
            element={
              <PublicRoute>
                <LoginPage />
              </PublicRoute>
            }
          />
          <Route
            path="/register"
            element={
              <PublicRoute>
                <RegisterPage />
              </PublicRoute>
            }
          />
          <Route
            path="/home"
            element={
              <GameRoute>
                <HomePage />
              </GameRoute>
            }
          />
          <Route
            path="/production"
            element={
              <GameRoute>
                <ProductionPage />
              </GameRoute>
            }
          />
          <Route
            path="/inventory"
            element={
              <GameRoute>
                <InventoryPage />
              </GameRoute>
            }
          />
          <Route
            path="/encyclopedia"
            element={
              <GameRoute>
                <EncyclopediaPage />
              </GameRoute>
            }
          />
          <Route
            path="/profile"
            element={
              <GameRoute>
                <ProfilePage />
              </GameRoute>
            }
          />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
