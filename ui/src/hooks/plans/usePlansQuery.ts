import { useQuery } from '@tanstack/react-query'
import { fetchPlans } from '@/api/plans'

export const usePlansQuery = (enabled = true) =>
  useQuery({
    queryKey: ['plans'],
    queryFn: fetchPlans,
    enabled,
  })
