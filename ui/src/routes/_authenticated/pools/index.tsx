import { createFileRoute } from '@tanstack/react-router'
import Pools from '@/features/pools'

export const Route = createFileRoute('/_authenticated/pools/')({
  component: Pools,
})
