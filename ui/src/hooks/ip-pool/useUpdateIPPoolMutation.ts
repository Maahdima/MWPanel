import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updateIPPool } from '@/api/ip-pool.ts'

export const useUpdateIPPoolMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: updateIPPool,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ip_pools_list'] })
    },
  })
}
