import { useQuery } from '@tanstack/react-query'
import { fetchPeerShareStatus } from '@/api/peers.ts'

export function usePeerShareQuery(id: number, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ['peer_share', id],
    queryFn: () => fetchPeerShareStatus(id),
    enabled: !!id && (options?.enabled ?? true),
  })
}
