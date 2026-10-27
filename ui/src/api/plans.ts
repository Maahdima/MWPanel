import { Plan, PlansResponseSchema, UpsertPlan } from '@/schema/plan'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchPlans = async (): Promise<Plan[]> => {
  const { data } = await axiosInstance.get('/plans')
  const parsed = PlansResponseSchema.parse(data)
  return parsed.data || []
}

export const createPlan = async (plan: UpsertPlan): Promise<void> => {
  await axiosInstance.post('/plans', plan)
}

export const updatePlan = async ({
  id,
  plan,
}: {
  id: number
  plan: UpsertPlan
}): Promise<void> => {
  await axiosInstance.put(`/plans/${id}`, plan)
}

export const deletePlan = async (id: number): Promise<void> => {
  await axiosInstance.delete(`/plans/${id}`)
}
