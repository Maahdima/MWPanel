import { useMutation } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { updateProfile } from '@/api/authentication.ts'
import { useAuthStore } from '@/stores/authStore.ts'

export const useUpdateProfileMutation = () => {
  const authStore = useAuthStore()
  const navigate = useNavigate()

  return useMutation({
    mutationFn: updateProfile,
    onSuccess: () => {
      authStore.auth.resetAccessToken()
      navigate({ to: '/sign-in' })
      toast.success('Account updated successfully.', {
        duration: 5000,
      })
    },
  })
}
