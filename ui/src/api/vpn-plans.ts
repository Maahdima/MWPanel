import {
  UpsertVpnPlan,
  VpnPlan,
  VpnPlansResponseSchema,
} from '@/schema/vpn-plan.ts'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchVpnPlans = async (): Promise<VpnPlan[]> => {
  const { data } = await axiosInstance.get('/plans')
  const parsed = VpnPlansResponseSchema.parse(data)
  return parsed.data || []
}

export const createVpnPlan = async (plan: UpsertVpnPlan): Promise<void> => {
  await axiosInstance.post('/plans', plan)
}

export const updateVpnPlan = async ({
  id,
  plan,
}: {
  id: number
  plan: UpsertVpnPlan
}): Promise<void> => {
  await axiosInstance.put(`/plans/${id}`, plan)
}

export const deleteVpnPlan = async (id: number): Promise<void> => {
  await axiosInstance.delete(`/plans/${id}`)
}
