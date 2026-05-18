import { BrowserRouter, Navigate, Outlet, Route, Routes, useLocation } from 'react-router-dom'
import './App.css'
import { getToken } from './api'
import { CandidateAppShell } from './CandidateAppShell'
import { JobDetailPage } from './JobDetailPage'
import { JobsPage } from './JobsPage'
import { LoginPage } from './LoginPage'
import { MyApplicationsPage } from './MyApplicationsPage'
import { ProfilePage } from './ProfilePage'
import { RegisterPage } from './RegisterPage'

function RequireCandidateAuth() {
  const loc = useLocation()
  if (!getToken()) {
    return <Navigate to="/login" replace state={{ from: loc.pathname }} />
  }
  return <Outlet />
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route element={<CandidateAppShell />}>
          <Route path="/jobs" element={<JobsPage />} />
          <Route path="/jobs/:id" element={<JobDetailPage />} />
          <Route element={<RequireCandidateAuth />}>
            <Route path="/my-applications" element={<MyApplicationsPage />} />
            <Route path="/profile" element={<ProfilePage />} />
          </Route>
        </Route>
        <Route path="/" element={<Navigate to="/jobs" replace />} />
        <Route path="*" element={<Navigate to="/jobs" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
