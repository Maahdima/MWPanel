import { useQuery } from '@tanstack/react-query'
import { fetchPeerQRCode } from '@/api/peers.ts'

export const usePeerQRCodeQuery = (
  uuid: string | undefined,
  options?: { enabled?: boolean }
) =>
  useQuery({
    queryKey: ['peer_qrcode', uuid],
    queryFn: () => fetchPeerQRCode(uuid),
    enabled: !!uuid && (options?.enabled ?? true),
  })
