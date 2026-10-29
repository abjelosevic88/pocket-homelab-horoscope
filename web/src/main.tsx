import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Navigate, Route, Routes, useLocation } from 'react-router-dom'
import './styles/base.css'
import './styles/horoscope.css'
import './styles/today.css'
import './styles/charts.css'
import './styles/sky.css'
import './styles/app.css'
import './styles/settings.css'
import { AppProvider, useApp } from './lib/app-context'
import { Loading, ToastProvider } from './components/ui'
import Layout from './components/Layout'
import { ErrorBoundary } from './components/ErrorBoundary'
import Login from './pages/Login'
import Today from './pages/Today'
import People from './pages/People'
import PersonPage from './pages/PersonPage'
import MatchPage from './pages/Match'
import Transits from './pages/Transits'
import Sky from './pages/Sky'
import Journal from './pages/Journal'
import Insights from './pages/Insights'
import Readings, { ReadingPage } from './pages/Readings'
import Settings from './pages/Settings'

function Gate({ children }: { children: React.ReactNode }) {
  const { status, user } = useApp()
  const loc = useLocation()
  if (!status) return <div className="boot"><img src="/favicon.svg" alt="" /><Loading /></div>
  if (status.needs_setup) return <Navigate to="/setup" replace />
  if (!user) return <Navigate to="/login" replace state={{ from: loc.pathname + loc.search }} />
  return <>{children}</>
}

function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login mode="login" />} />
      <Route path="/setup" element={<Login mode="setup" />} />
      <Route path="/settings/:tab?" element={<Gate><ErrorBoundary where="settings"><Settings /></ErrorBoundary></Gate>} />
      <Route element={<Gate><ErrorBoundary where="horoscope"><Layout /></ErrorBoundary></Gate>}>
        <Route path="/" element={<Today />} />
        <Route path="/people" element={<People />} />
        <Route path="/people/:id/:tab?" element={<PersonPage />} />
        <Route path="/match" element={<MatchPage />} />
        <Route path="/transits" element={<Transits />} />
        <Route path="/sky/:view?" element={<Sky />} />
        <Route path="/journal" element={<Journal />} />
        <Route path="/journal/insights" element={<Insights />} />
        <Route path="/readings" element={<Readings />} />
        <Route path="/readings/:id" element={<ReadingPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <ErrorBoundary where="app"><ToastProvider><AppProvider><App /></AppProvider></ToastProvider></ErrorBoundary>
    </BrowserRouter>
  </React.StrictMode>,
)
