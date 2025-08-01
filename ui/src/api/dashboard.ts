import {
  DailyTrafficUsage,
  dailyTrafficUsageResponseSchema,
  DeviceData,
  deviceDataResponseSchema,
} from '@/schema/dashboard.ts'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchDeviceData = async (): Promise<DeviceData> => {
  const { data } = await axiosInstance.get('/device/stats')
  const parsed = deviceDataResponseSchema.parse(data)
  return parsed.data
}

export const fetchDailyTrafficUsage = async (
  range: number = 7
): Promise<DailyTrafficUsage> => {
  const { data } = await axiosInstance.get('/device/traffic', {
    params: { range },
  })
  const parsed = dailyTrafficUsageResponseSchema.parse(data)
  return parsed.data
}
