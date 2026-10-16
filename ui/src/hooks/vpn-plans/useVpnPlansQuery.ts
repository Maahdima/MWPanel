import { useQuery } from '@tanstack/react-query'
import { fetchVpnPlans } from '@/api/vpn-plans.ts'

export const useVpnPlansQuery = (enabled = true) =>
  useQuery({
    queryKey: ['vpn_plans'],
    queryFn: fetchVpnPlans,
    enabled,
  })
