import { useMutation } from '@tanstack/react-query'
import { fetchPeerAllowedAddress } from '@/api/peers.ts'

export const usePeerAllowedAddressMutation = () => {
  return useMutation({
    mutationFn: fetchPeerAllowedAddress,
  })
}
