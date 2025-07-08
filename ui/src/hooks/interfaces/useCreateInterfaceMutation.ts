import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createInterface } from '@/api/interfaces.ts'

export const useCreateInterfaceMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createInterface,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interfaces_list'] })
    },
  })
}
