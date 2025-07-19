import { useQuery } from '@tanstack/react-query'
import { fetchPeerShare } from '@/api/peers.ts'

export function usePeerShareQuery(
  uuid: string,
  options?: { enabled?: boolean }
) {
  return useQuery({
    queryKey: ['peer_share', uuid],
    queryFn: () => fetchPeerShare(uuid),
    enabled: !!uuid && (options?.enabled ?? true),
  })
}
