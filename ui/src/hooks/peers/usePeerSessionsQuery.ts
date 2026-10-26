import { useQuery } from '@tanstack/react-query'
import { fetchPeerSessions } from '@/api/peers.ts'

export function usePeerSessionsQuery(
  id: number,
  options?: { enabled?: boolean; scope?: 'admin' | 'reseller' }
) {
  return useQuery({
    queryKey: ['peer_sessions', id, options?.scope ?? 'admin'],
    queryFn: () => fetchPeerSessions(id, options?.scope),
    enabled: !!id && (options?.enabled ?? true),
    refetchInterval: options?.enabled === false ? false : 10_000,
  })
}
