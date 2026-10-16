import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchShopSales, updateShopSales } from '@/api/shop-sales.ts'

export const useShopSalesQuery = () =>
  useQuery({
    queryKey: ['shop_sales'],
    queryFn: fetchShopSales,
  })

export const useUpdateShopSalesMutation = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updateShopSales,
    onSuccess: (sales) => {
      queryClient.setQueryData(['shop_sales'], sales)
    },
  })
}
