import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createServer } from '@/api/servers.ts'

export const useCreateServerMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createServer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers_list'] })
    },
  })
}
