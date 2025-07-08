import { useQuery } from '@tanstack/react-query'
import { fetchPeersList } from '@/api/peers.ts'

export const usePeersListQuery = () =>
  useQuery({
    queryKey: ['peers_list'],
    queryFn: fetchPeersList,
  })
