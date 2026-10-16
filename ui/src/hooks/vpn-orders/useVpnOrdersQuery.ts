import { useQuery } from '@tanstack/react-query'
import { fetchVpnOrders } from '@/api/vpn-orders.ts'

export const useVpnOrdersQuery = (enabled = true) =>
  useQuery({
    queryKey: ['vpn_orders'],
    queryFn: fetchVpnOrders,
    enabled,
  })
