import { useMutation, useQueryClient } from '@tanstack/react-query'
import { syncPeers } from '@/api/peers.ts'

export const useSyncPeersMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: syncPeers,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['peers_list'] })
    },
  })
}
