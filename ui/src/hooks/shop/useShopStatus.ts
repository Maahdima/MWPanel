import { useQuery } from '@tanstack/react-query'
import { fetchShopStatus } from '@/api/shop-status.ts'

export const useShopStatusQuery = () =>
  useQuery({
    queryKey: ['shop_status'],
    queryFn: fetchShopStatus,
    retry: false,
    staleTime: 30_000,
  })

export const useShopEnabled = () => {
  const query = useShopStatusQuery()
  return {
    enabled: query.data?.enabled === true,
    isLoading: query.isLoading,
  }
}
