import { useQuery } from '@tanstack/react-query'
import { fetchPeerQRCode } from '@/api/peers.ts'

export const usePeerQRCodeQuery = (
  id: number,
  options?: { enabled?: boolean }
) =>
  useQuery({
    queryKey: ['peer_qrcode', id],
    queryFn: () => fetchPeerQRCode(id),
    enabled: !!id && (options?.enabled ?? true),
  })
