import { useQuery } from '@tanstack/react-query'
import { fetchDailyTrafficUsage } from '@/api/dashboard.ts'

export const useDailyTrafficUsageQuery = (range: number = 90) =>
  useQuery({
    queryKey: ['daily_traffic_usage', range],
    queryFn: () => fetchDailyTrafficUsage(range),
  })
