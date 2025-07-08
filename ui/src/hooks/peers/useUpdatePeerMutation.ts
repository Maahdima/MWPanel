import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updatePeer } from '@/api/peers.ts'

export const useUpdatePeerMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: updatePeer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['peers_list'] })
    },
  })
}
