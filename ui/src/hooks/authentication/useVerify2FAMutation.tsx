import { useMutation, useQueryClient } from '@tanstack/react-query'
import { verify2FA } from '@/api/authentication.ts'

export const useVerify2FAMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: verify2FA,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['device_data'] })
    },
  })
}
