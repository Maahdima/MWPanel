import { useQuery } from '@tanstack/react-query'
import { fetchDeviceData } from '@/api/dashboard.ts'

export const useDeviceDataQuery = () =>
  useQuery({
    queryKey: ['device_data'],
    queryFn: fetchDeviceData,
  })
