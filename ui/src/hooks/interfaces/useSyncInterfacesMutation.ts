import { useMutation, useQueryClient } from '@tanstack/react-query'
import { syncInterfaces } from '@/api/interfaces.ts'

export const useSyncInterfacesMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: syncInterfaces,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interfaces_list'] })
    },
  })
}
