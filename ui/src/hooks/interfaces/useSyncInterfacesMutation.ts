import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { syncInterfaces } from '@/api/interfaces.ts'

export const useSyncInterfacesMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: syncInterfaces,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interfaces_list'] })
      toast.success('Interfaces synced successfully', {
        duration: 5000,
      })
    },
  })
}
