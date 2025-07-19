import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updatePeerShareStatus } from '@/api/peers.ts'

export const useUpdatePeerShareStatusMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: updatePeerShareStatus,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['peer_share'] })
    },
  })
}
