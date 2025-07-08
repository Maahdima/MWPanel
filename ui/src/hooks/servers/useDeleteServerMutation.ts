import { useMutation, useQueryClient } from '@tanstack/react-query'
import { deleteServer } from '@/api/servers.ts'

export const useDeleteServerMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deleteServer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers_list'] })
    },
  })
}
