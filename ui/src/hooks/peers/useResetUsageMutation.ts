import { useMutation, useQueryClient } from '@tanstack/react-query'
import { resetPeerUsage } from '@/api/peers.ts'

export const useResetUsageMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: resetPeerUsage,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['peers_list'] })
    },
  })
}
