import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { syncPeers } from '@/api/peers.ts'

export const useSyncPeersMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: syncPeers,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['peers_list'] })
      toast.success('Peers synced successfully', {
        duration: 5000,
      })
    },
  })
}
