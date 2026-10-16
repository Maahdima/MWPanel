import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  createVpnPlan,
  deleteVpnPlan,
  updateVpnPlan,
} from '@/api/vpn-plans.ts'

export const useCreateVpnPlanMutation = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createVpnPlan,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vpn_plans'] })
    },
  })
}

export const useUpdateVpnPlanMutation = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updateVpnPlan,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vpn_plans'] })
    },
  })
}

export const useDeleteVpnPlanMutation = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteVpnPlan,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vpn_plans'] })
    },
  })
}
