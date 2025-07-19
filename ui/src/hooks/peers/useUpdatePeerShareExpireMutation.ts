import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updatePeerShareExpire } from '@/api/peers.ts'

export const useUpdatePeerShareExpireMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: updatePeerShareExpire,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['peer_share'] })
    },
  })
}
