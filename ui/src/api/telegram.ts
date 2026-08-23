import {
  PeerTelegramStatus,
  PeerTelegramStatusResponseSchema,
  TelegramStatus,
  TelegramStatusResponseSchema,
} from '@/schema/telegram.ts'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchTelegramStatus = async (): Promise<TelegramStatus> => {
  const { data } = await axiosInstance.get('/telegram/status')
  const parsed = TelegramStatusResponseSchema.parse(data)
  return parsed.data
}

export const fetchPeerTelegramStatus = async (
  uuid: string | undefined
): Promise<PeerTelegramStatus> => {
  const { data } = await axiosInstance.get(`/user/${uuid}/telegram`)
  const parsed = PeerTelegramStatusResponseSchema.parse(data)
  return parsed.data
}
