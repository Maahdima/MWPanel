import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updatePeerStatus } from '@/api/peers.ts'

export const useUpdatePeerStatusMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: updatePeerStatus,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['peers_list'] })
      queryClient.invalidateQueries({ queryKey: ['device_data'] })
    },
  })
}
