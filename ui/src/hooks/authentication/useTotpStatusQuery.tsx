import { useQuery } from '@tanstack/react-query'
import { getTotpStatus } from '@/api/authentication.ts'

export const useTotpStatusQuery = () =>
  useQuery({
    queryKey: ['totp_status'],
    queryFn: getTotpStatus,
  })
