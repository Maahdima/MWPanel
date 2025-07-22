import { useQuery } from '@tanstack/react-query'
import { fetchUserDetails } from '@/api/user.ts'

export const useUserDetailsQuery = (uuid: string | undefined) =>
  useQuery({
    queryKey: ['user_details', uuid],
    queryFn: () => fetchUserDetails(uuid),
    enabled: !!uuid,
  })
