import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import './App.css'
import { getToken } from './api'
import { AiAssistantPage } from './AiAssistantPage'
import { HrAppShell } from './HrAppShell'
import { JobApplications } from './JobApplications'
import { ApplicationsByJobPage } from './ApplicationsByJobPage'
import { JobsPage } from './JobsPage'
import { LoginPage } from './LoginPage'
import { RegisterPage } from './RegisterPage'

function RequireHrAuth() {
  if (!getToken()) return <Navigate to="/login" replace />
  return <HrAppShell />
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route element={<RequireHrAuth />}>
          <Route index element={<JobsPage />} />
          <Route path="applications" element={<ApplicationsByJobPage />} />
          <Route path="jobs/:id" element={<JobApplications />} />
          <Route path="ai" element={<AiAssistantPage />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
