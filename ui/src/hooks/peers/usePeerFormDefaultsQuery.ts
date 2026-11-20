import { useQuery } from '@tanstack/react-query'
import { fetchPeerFormDefaults } from '@/api/peer-defaults'
import { defaultExpireTimeFromDays } from '@/lib/peer-default-expire'

export const usePeerFormDefaultsQuery = () =>
  useQuery({
    queryKey: ['peer_form_defaults'],
    queryFn: fetchPeerFormDefaults,
    staleTime: 5 * 60 * 1000,
  })

export function useDefaultPeerExpireTime(): string | null | undefined {
  const { data } = usePeerFormDefaultsQuery()
  if (!data) {
    return undefined
  }

  return defaultExpireTimeFromDays(data.default_expire_days)
}
