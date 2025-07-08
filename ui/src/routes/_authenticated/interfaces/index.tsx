import { createFileRoute } from '@tanstack/react-router'
import Interfaces from '@/features/interfaces'

export const Route = createFileRoute('/_authenticated/interfaces/')({
  component: Interfaces,
})
