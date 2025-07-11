import { useQuery } from '@tanstack/react-query'
import { fetchPeerDetails } from '@/api/peers.ts'

export const usePeerStatsQuery = (uuid: string | undefined) =>
  useQuery({
    queryKey: ['peer_details', uuid],
    queryFn: () => fetchPeerDetails(uuid),
    enabled: !!uuid,
  })
