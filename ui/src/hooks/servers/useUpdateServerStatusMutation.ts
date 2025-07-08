import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updateServerStatus } from '@/api/servers.ts'

export const useUpdateServerStatusMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: updateServerStatus,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers_list'] })
    },
  })
}
