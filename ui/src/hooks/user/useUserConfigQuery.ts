import { useQuery } from '@tanstack/react-query'
import { fetchUserConfig } from '@/api/user.ts'

export const useUserConfigQuery = (
  uuid: string | undefined,
  options?: { enabled?: boolean }
) =>
  useQuery({
    queryKey: ['user_config', uuid],
    queryFn: () => fetchUserConfig(uuid),
    enabled: !!uuid && (options?.enabled ?? true),
  })
