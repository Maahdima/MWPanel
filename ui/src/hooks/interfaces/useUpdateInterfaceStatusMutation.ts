import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updateInterfaceStatus } from '@/api/interfaces.ts'

export const useUpdateInterfaceStatusMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: updateInterfaceStatus,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interfaces_list'] })
    },
  })
}
