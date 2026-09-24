import { lazy, Suspense, type ReactNode } from 'react'
import { Navigate, Outlet, createBrowserRouter } from 'react-router-dom'
import { AppShell } from '@/components/common/AppShell'
import { useAuthStore } from '@/stores/auth'
import { LoginPage } from '@/pages/LoginPage'

const DiversPage = lazy(() => import('@/pages/DiversPage').then((module) => ({ default: module.DiversPage })))
const PlansPage = lazy(() => import('@/pages/PlansPage').then((module) => ({ default: module.PlansPage })))
const ExposuresPage = lazy(() => import('@/pages/ExposuresPage').then((module) => ({ default: module.ExposuresPage })))
const AssessmentsPage = lazy(() => import('@/pages/AssessmentsPage').then((module) => ({ default: module.AssessmentsPage })))
const AuditPage = lazy(() => import('@/pages/AuditPage').then((module) => ({ default: module.AuditPage })))

const deferred = (page: ReactNode) => <Suspense fallback={<div className="loading-state">Loading workspace…</div>}>{page}</Suspense>

function RequireAuth() {
  return useAuthStore((state) => state.user) ? <Outlet /> : <Navigate to="/login" replace />
}

function RequireSupervisor() {
  const role = useAuthStore((state) => state.user?.role)
  return role === 'supervisor' || role === 'admin' ? <Outlet /> : <Navigate to="/assessments" replace />
}

export const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  {
    element: <RequireAuth />,
    children: [{
      element: <AppShell />,
      children: [
        { index: true, element: <Navigate to="/plans" replace /> },
        { path: '/divers', element: deferred(<DiversPage />) },
        { path: '/plans', element: deferred(<PlansPage />) },
        { path: '/exposures', element: deferred(<ExposuresPage />) },
        { path: '/assessments', element: deferred(<AssessmentsPage />) },
        { element: <RequireSupervisor />, children: [{ path: '/audit', element: deferred(<AuditPage />) }] },
      ],
    }],
  },
  { path: '*', element: <Navigate to="/plans" replace /> },
])
