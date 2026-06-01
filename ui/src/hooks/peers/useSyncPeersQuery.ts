import { useQuery } from '@tanstack/react-query'
import { fetchSyncPeers } from '@/api/peers.ts'

export const useSyncPeersQuery = (enabled: boolean, interfaceName?: string) =>
  useQuery({
	queryKey: ['sync_peers', interfaceName ?? ''],
	queryFn: () => fetchSyncPeers(interfaceName),
	enabled,
  })

