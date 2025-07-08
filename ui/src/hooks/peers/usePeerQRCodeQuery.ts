import { useQuery } from '@tanstack/react-query'
import { fetchPeerQRCode } from '@/api/peers.ts'

export const usePeerQRCodeQuery = (id: number) =>
  useQuery({
    queryKey: ['peer_qrcode', id],
    queryFn: () => fetchPeerQRCode(id),
    enabled: !!id,
  })
