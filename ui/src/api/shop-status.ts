import { ShopStatus, ShopStatusResponseSchema } from '@/schema/shop-status.ts'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchShopStatus = async (): Promise<ShopStatus> => {
  const { data } = await axiosInstance.get('/shop/status')
  return ShopStatusResponseSchema.parse(data).data
}
