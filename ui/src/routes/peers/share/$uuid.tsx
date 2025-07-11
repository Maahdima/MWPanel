import { createFileRoute } from '@tanstack/react-router'
import PeerShare from '@/features/share'

export const Route = createFileRoute('/peers/share/$uuid')({
  component: PeerShare,
})
