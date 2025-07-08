import { createFileRoute } from '@tanstack/react-router'
import Peers from '@/features/peers'

export const Route = createFileRoute('/_authenticated/peers/')({
  component: Peers,
})
