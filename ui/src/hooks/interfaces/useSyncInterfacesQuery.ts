import { useQuery } from '@tanstack/react-query'
import { fetchSyncInterfaces } from '@/api/interfaces.ts'

export const useSyncInterfacesQuery = (enabled: boolean, serverID?: number) =>
  useQuery({
    queryKey: ['sync_interfaces', serverID],
    queryFn: () => fetchSyncInterfaces(serverID),
    enabled: enabled && !!serverID,
  })
