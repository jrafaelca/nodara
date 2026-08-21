import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { useAuth } from '@/lib/auth.jsx'
import ChangePasswordPage from '@/pages/change-password.jsx'
import DashboardPage from '@/pages/dashboard.jsx'
import ForgotPasswordPage from '@/pages/forgot-password.jsx'
import LoginPage from '@/pages/login.jsx'
import ResetPasswordPage from '@/pages/reset-password.jsx'

function LoadingScreen() {
  return <main className="flex min-h-svh items-center justify-center bg-background text-sm text-muted-foreground">Loading console…</main>
}

function ProtectedRoute({ children }) {
  const { user, loading } = useAuth()
  if (loading) return <LoadingScreen />
  if (!user) return <Navigate to="/login" replace />
  if (user.password_change_required) return <Navigate to="/change-password" replace />
  return children
}

function GuestRoute({ children }) {
  const { user, loading } = useAuth()
  if (loading) return <LoadingScreen />
  if (!user) return children
  return <Navigate to={user.password_change_required ? '/change-password' : '/'} replace />
}

function ChangePasswordRoute() {
  const { user, loading } = useAuth()
  if (loading) return <LoadingScreen />
  if (!user) return <Navigate to="/login" replace />
  return <ChangePasswordPage />
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />
        <Route path="/login" element={<GuestRoute><LoginPage /></GuestRoute>} />
        <Route path="/forgot-password" element={<GuestRoute><ForgotPasswordPage /></GuestRoute>} />
        <Route path="/reset-password" element={<GuestRoute><ResetPasswordPage /></GuestRoute>} />
        <Route path="/change-password" element={<ChangePasswordRoute />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
