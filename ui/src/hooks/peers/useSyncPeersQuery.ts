import { useQuery } from '@tanstack/react-query'
import { fetchSyncPeers } from '@/api/peers.ts'

export const useSyncPeersQuery = (
  enabled: boolean,
  interfaceName?: string,
  serverID?: number
) =>
  useQuery({
    queryKey: ['sync_peers', serverID, interfaceName ?? ''],
    queryFn: () => fetchSyncPeers(interfaceName, serverID),
    enabled: enabled && !!serverID,
  })
