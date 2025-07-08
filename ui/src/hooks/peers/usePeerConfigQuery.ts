import { useQuery } from '@tanstack/react-query'
import { fetchPeerConfig } from '@/api/peers.ts'

export const usePeerConfigQuery = (id: number) =>
  useQuery({
    queryKey: ['peer_config', id],
    queryFn: () => fetchPeerConfig(id),
    enabled: !!id,
  })
