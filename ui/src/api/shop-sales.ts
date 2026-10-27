import { ShopSales, ShopSalesResponseSchema } from '@/schema/shop-sale'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchShopSales = async (): Promise<ShopSales> => {
  const { data } = await axiosInstance.get('/shop/sales')
  return ShopSalesResponseSchema.parse(data).data
}

export const updateShopSales = async (sales: ShopSales): Promise<ShopSales> => {
  const { data } = await axiosInstance.put('/shop/sales', sales)
  return ShopSalesResponseSchema.parse(data).data
}
