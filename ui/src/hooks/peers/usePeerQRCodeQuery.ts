import { useQuery } from '@tanstack/react-query'
import { fetchPeerQRCode } from '@/api/peers.ts'

export const usePeerQRCodeQuery = (uuid: string | undefined) =>
  useQuery({
    queryKey: ['peer_qrcode', uuid],
    queryFn: () => fetchPeerQRCode(uuid),
    enabled: !!uuid,
  })
