import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createIPPool } from '@/api/ip-pool.ts'

export const useCreateIPPoolMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createIPPool,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ip_pools_list'] })
    },
  })
}
