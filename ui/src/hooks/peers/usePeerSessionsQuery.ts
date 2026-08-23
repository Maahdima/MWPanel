import { useQuery } from '@tanstack/react-query'
import { fetchPeerSessions } from '@/api/peers.ts'

export function usePeerSessionsQuery(
  id: number,
  options?: { enabled?: boolean }
) {
  return useQuery({
    queryKey: ['peer_sessions', id],
    queryFn: () => fetchPeerSessions(id),
    enabled: !!id && (options?.enabled ?? true),
    refetchInterval: options?.enabled === false ? false : 10_000,
  })
}
