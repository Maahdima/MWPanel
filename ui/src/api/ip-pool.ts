import {
  CreateIPPoolRequest,
  CreateIPPoolSchema,
  IPPool,
  IPPoolResponseSchema,
  IPPoolsResponseSchema,
  UpdateIPPoolRequest,
} from '@/schema/ip-pool.ts'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchIPPoolsList = async (): Promise<IPPool[]> => {
  const { data } = await axiosInstance.get('/ip-pool')
  const parsed = IPPoolsResponseSchema.parse(data)
  return parsed.data || []
}

export const createIPPool = async (
  ipPool: CreateIPPoolRequest
): Promise<IPPool> => {
  const validated = CreateIPPoolSchema.parse(ipPool)
  const { data } = await axiosInstance.post('/ip-pool', validated)
  const parsed = IPPoolResponseSchema.parse(data)
  return parsed.data
}

export const updateIPPool = async (
  ipPool: UpdateIPPoolRequest
): Promise<IPPool> => {
  const { data } = await axiosInstance.put(`/ip-pool/${ipPool.id}`, ipPool)
  const parsed = IPPoolResponseSchema.parse(data)
  return parsed.data
}

export const deleteIPPool = async (id: number): Promise<void> => {
  await axiosInstance.delete(`/ip-pool/${id}`)
}
