import { useQuery } from '@tanstack/react-query'
import { fetchSyncInterfaces } from '@/api/interfaces.ts'

export const useSyncInterfacesQuery = (enabled: boolean) =>
  useQuery({
    queryKey: ['sync_interfaces'],
    queryFn: fetchSyncInterfaces,
    enabled,
  })

