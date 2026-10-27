import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createPlan, deletePlan, updatePlan } from '@/api/plans'

export const useCreatePlanMutation = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createPlan,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plans'] })
    },
  })
}

export const useUpdatePlanMutation = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updatePlan,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plans'] })
    },
  })
}

export const useDeletePlanMutation = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deletePlan,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plans'] })
    },
  })
}
