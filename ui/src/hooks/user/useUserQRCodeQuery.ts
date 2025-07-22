import { useQuery } from '@tanstack/react-query'
import { fetchUserQRCode } from '@/api/user.ts'

export const useUserQRCodeQuery = (
  uuid: string | undefined,
  options?: { enabled?: boolean }
) =>
  useQuery({
    queryKey: ['user_qrcode', uuid],
    queryFn: () => fetchUserQRCode(uuid),
    enabled: !!uuid && (options?.enabled ?? true),
  })
