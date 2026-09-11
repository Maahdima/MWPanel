import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  confirmTotpSetup,
  disableTotp,
  startTotpSetup,
} from '@/api/authentication.ts'

export const useStartTotpSetupMutation = () =>
  useMutation({
    mutationFn: startTotpSetup,
  })

export const useConfirmTotpSetupMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: confirmTotpSetup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['totp_status'] })
      toast.success('Two-factor authentication enabled.')
    },
  })
}

export const useDisableTotpMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: disableTotp,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['totp_status'] })
      toast.success('Two-factor authentication disabled.')
    },
  })
}
