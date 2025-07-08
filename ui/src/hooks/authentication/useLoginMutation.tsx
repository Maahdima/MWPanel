import { useMutation, useQueryClient } from '@tanstack/react-query'
import { login } from '@/api/authentication.ts'

export const useLoginMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: login,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['device_data'] })
    },
  })
}
