import { createFileRoute } from '@tanstack/react-router'
import Servers from '@/features/servers'

export const Route = createFileRoute('/_authenticated/servers/')({
  component: Servers,
})
