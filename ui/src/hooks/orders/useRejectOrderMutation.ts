import { useMutation, useQueryClient } from '@tanstack/react-query'
import { rejectOrder } from '@/api/orders'

export const useRejectOrderMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: rejectOrder,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] })
    },
  })
}
