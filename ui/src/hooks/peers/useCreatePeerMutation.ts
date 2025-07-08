import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createPeer } from '@/api/peers.ts'

export const useCreatePeerMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createPeer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['peers_list'] })
    },
  })
}
