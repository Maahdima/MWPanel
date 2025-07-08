import { useMutation, useQueryClient } from '@tanstack/react-query'
import { deleteInterface } from '@/api/interfaces.ts'

export const useDeleteInterfaceMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deleteInterface,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interfaces_list'] })
    },
  })
}
