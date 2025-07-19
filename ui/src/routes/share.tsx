import { z } from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import PeerShare from '@/features/share'

export const Route = createFileRoute('/share')({
  component: PeerShare,

  validateSearch: z.object({
    shareId: z.string().uuid(),
  }),
})
