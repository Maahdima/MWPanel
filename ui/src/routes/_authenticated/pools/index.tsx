import { createFileRoute } from '@tanstack/react-router'
import Pools from '@/features/ip-pools'

export const Route = createFileRoute('/_authenticated/pools/')({
  component: Pools,
})
