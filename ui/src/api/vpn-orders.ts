import { VpnOrder, VpnOrdersResponseSchema } from '@/schema/vpn-order.ts'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchVpnOrders = async (): Promise<VpnOrder[]> => {
  const { data } = await axiosInstance.get('/orders')
  const parsed = VpnOrdersResponseSchema.parse(data)
  return parsed.data || []
}

export const acceptVpnOrder = async (id: number): Promise<void> => {
  await axiosInstance.post(`/orders/${id}/accept`)
}
