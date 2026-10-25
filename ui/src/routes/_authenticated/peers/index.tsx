import { createFileRoute } from '@tanstack/react-router'
import { usePrincipal } from '@/hooks/resellers/useResellerQueries'
import Peers from '@/features/peers'
import { MyPeers } from '@/features/resellers/my-peers'

export const Route = createFileRoute('/_authenticated/peers/')({
  component: PeerPage,
})
function PeerPage() {
  const { data } = usePrincipal()
  return data?.role === 'reseller' ? <MyPeers /> : <Peers />
}
