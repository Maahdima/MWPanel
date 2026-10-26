import { useQuery } from '@tanstack/react-query'
import { fetchShopStatus } from '@/api/shop-status.ts'

export const useShopStatusQuery = (enabled = true) =>
  useQuery({
    queryKey: ['shop_status'],
    enabled,
    queryFn: fetchShopStatus,
    retry: false,
    staleTime: 30_000,
  })

export const useShopEnabled = (enabled = true) => {
  const query = useShopStatusQuery(enabled)
  return {
    enabled: query.data?.enabled === true,
    isLoading: query.isLoading,
  }
}
