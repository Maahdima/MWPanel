import { useMutation, useQueryClient } from '@tanstack/react-query'
import { deleteIPPool } from '@/api/ip-pool.ts'

export const useDeleteIPPoolMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deleteIPPool,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ip_pools_list'] })
    },
  })
}
