import { useQuery } from '@tanstack/react-query'
import { fetchPeerConfig } from '@/api/peers.ts'

export const usePeerConfigQuery = (
  uuid: string | undefined,
  options?: { enabled?: boolean }
) =>
  useQuery({
    queryKey: ['peer_config', uuid],
    queryFn: () => fetchPeerConfig(uuid),
    enabled: !!uuid && (options?.enabled ?? true),
  })
