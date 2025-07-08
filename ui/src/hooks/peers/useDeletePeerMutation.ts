import { useMutation, useQueryClient } from '@tanstack/react-query'
import { deletePeer } from '@/api/peers.ts'

export const useDeletePeerMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deletePeer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['peers_list'] })
    },
  })
}
