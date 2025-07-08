import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updateInterface } from '@/api/interfaces.ts'

export const useUpdateInterfaceMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: updateInterface,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interfaces_list'] })
    },
  })
}
