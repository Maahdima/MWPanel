import { useMutation, useQueryClient } from '@tanstack/react-query'
import { resetPeerUsages } from '@/api/peers.ts'

export const useResetUsagesMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: resetPeerUsages,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['peers_list'] })
    },
  })
}
