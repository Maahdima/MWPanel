import { createFileRoute } from '@tanstack/react-router'
import ResellersPage from '@/features/resellers'

export const Route = createFileRoute('/_authenticated/resellers/')({
  component: ResellersPage,
})
