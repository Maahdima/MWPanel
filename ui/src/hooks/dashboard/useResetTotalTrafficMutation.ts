import { useMutation, useQueryClient } from '@tanstack/react-query'
import { resetTotalTrafficUsage } from '@/api/dashboard.ts'

export const useResetTotalTrafficMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: resetTotalTrafficUsage,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['device_data'] })
    },
  })
}
