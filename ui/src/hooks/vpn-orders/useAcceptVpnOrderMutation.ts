import { useMutation, useQueryClient } from '@tanstack/react-query'
import { acceptVpnOrder } from '@/api/vpn-orders.ts'

export const useAcceptVpnOrderMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: acceptVpnOrder,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vpn_orders'] })
    },
  })
}
