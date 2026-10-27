import { useQuery } from '@tanstack/react-query'
import { fetchOrders } from '@/api/orders'

export const useOrdersQuery = (enabled = true) =>
  useQuery({
    queryKey: ['orders'],
    queryFn: fetchOrders,
    enabled,
  })
