import { useQuery } from '@tanstack/react-query'
import { fetchPeerTelegramStatus } from '@/api/telegram.ts'

export const usePeerTelegramStatusQuery = (uuid: string | undefined) =>
  useQuery({
    queryKey: ['peer_telegram_status', uuid],
    queryFn: () => fetchPeerTelegramStatus(uuid),
    enabled: !!uuid,
    refetchInterval: 8_000,
  })
