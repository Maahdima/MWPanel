import { Order, OrdersResponseSchema } from '@/schema/order'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchOrders = async (): Promise<Order[]> => {
  const { data } = await axiosInstance.get('/orders')
  const parsed = OrdersResponseSchema.parse(data)
  return parsed.data || []
}

export const acceptOrder = async (id: number): Promise<void> => {
  await axiosInstance.post(`/orders/${id}/accept`)
}
