import { useQuery } from '@tanstack/react-query'
import { fetchDeviceData } from '@/api/dashboard.ts'

export const useDeviceDataQuery = (serverID?: number) =>
  useQuery({
    queryKey: ['device_data', serverID],
    queryFn: () => fetchDeviceData(serverID),
  })
