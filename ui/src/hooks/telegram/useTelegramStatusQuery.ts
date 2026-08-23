import { useQuery } from '@tanstack/react-query'
import { fetchTelegramStatus } from '@/api/telegram.ts'

export const useTelegramStatusQuery = () =>
  useQuery({
    queryKey: ['telegram_status'],
    queryFn: fetchTelegramStatus,
    staleTime: 60_000,
    retry: false,
  })
