import './App.css'
import AuthModal from './components/AuthModal'
import { Link, Route } from 'wouter'
import ResourcesPage from './pages/ResourcesPage'
import ProductionPage from './pages/ProductionPage'
import ProductionDetailPage from './pages/ProductionDetailPage'
import SimulationPage from './pages/SimulationPage'
import { useAuth } from './contexts/AuthContext'

function App() {
  const { token } = useAuth()

  // If no token, render only the auth page — prevent access to any route.
  if (!token) return <AuthModal />

  return (
    <>
      <header style={{padding: 12, borderBottom: '1px solid #eee', display: 'flex', gap: 12}}>
        <nav style={{display: 'flex', gap: 12}}>
          <Link href="/resources">Resources</Link>
          <Link href="/production">Production</Link>
        </nav>
      </header>
      <main style={{padding: 16}}>
        <Route path="/resources">
          <ResourcesPage />
        </Route>
        <Route path="/production">
          <ProductionPage />
        </Route>
        <Route path="/production/:id">
          {(params: any) => <ProductionDetailPage />}
        </Route>
        <Route path="/simulate/:pid">
          {(params: any) => <SimulationPage />}
        </Route>
        <Route path="/">
          <ResourcesPage />
        </Route>
      </main>
    </>
  )
}

export default App
