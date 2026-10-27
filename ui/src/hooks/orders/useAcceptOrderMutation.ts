import { useMutation, useQueryClient } from '@tanstack/react-query'
import { acceptOrder } from '@/api/orders'

export const useAcceptOrderMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: acceptOrder,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] })
    },
  })
}
