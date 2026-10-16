import { createFileRoute } from '@tanstack/react-router'
import Plans from '@/features/plans'

export const Route = createFileRoute('/_authenticated/plans/')({
  component: Plans,
})
