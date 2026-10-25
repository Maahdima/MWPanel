import { createFileRoute, Navigate } from '@tanstack/react-router'
import { usePrincipal } from '@/hooks/resellers/useResellerQueries'
import Dashboard from '@/features/dashboard'

export const Route = createFileRoute('/_authenticated/')({
  component: HomePage,
})
function HomePage() {
  const { data } = usePrincipal()
  if (!data) return null
  return data.role === 'reseller' ? (
    <Navigate to='/billing' replace />
  ) : (
    <Dashboard />
  )
}
