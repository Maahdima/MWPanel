import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updateServer } from '@/api/servers.ts'

export const useUpdateServerMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: updateServer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers_list'] })
    },
  })
}
