import { PeerStats, PeerStatsResponseSchema } from '@/schema/peers.ts'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchUserQRCode = async (
  uuid: string | undefined
): Promise<string> => {
  const response = await axiosInstance.get(`/user/${uuid}/qrcode`, {
    responseType: 'blob',
  })

  return URL.createObjectURL(response.data)
}

export const fetchUserConfig = async (
  uuid: string | undefined
): Promise<string> => {
  const response = await axiosInstance.get(`/user/${uuid}/config`, {
    responseType: 'blob',
  })

  return response.data
}

export const fetchUserDetails = async (
  uuid: string | undefined
): Promise<PeerStats> => {
  const { data } = await axiosInstance.get(`/user/${uuid}/details`)
  const parsed = PeerStatsResponseSchema.parse(data)
  return parsed.data
}
