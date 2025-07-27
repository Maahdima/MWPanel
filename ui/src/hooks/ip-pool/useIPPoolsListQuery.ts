import { useQuery } from '@tanstack/react-query'
import { fetchIPPoolsList } from '@/api/ip-pool.ts'

export const useIPPoolsListQuery = () =>
  useQuery({
    queryKey: ['ip_pools_list'],
    queryFn: fetchIPPoolsList,
  })
