import { z } from 'zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createApiResponseSchema } from '@/schema/api-response'
import {
  BillingBookSchema,
  InstructionsSchema,
  PrincipalSchema,
  ReceiptSchema,
  ResellerSchema,
} from '@/schema/billing'
import { toast } from 'sonner'
import axiosInstance from '@/api/axios-instance'
import { useAuthStore } from '@/stores/authStore'

async function read<T extends z.ZodTypeAny>(
  path: string,
  schema: T
): Promise<z.infer<T>> {
  const { data } = await axiosInstance.get(path)
  return createApiResponseSchema(schema).parse(data).data
}
export const usePrincipal = () => {
  const token = useAuthStore((state) => state.auth.accessToken)
  return useQuery({
    queryKey: ['principal'],
    queryFn: () => read('/auth/me', PrincipalSchema),
    enabled: !!token,
    staleTime: 10000,
    retry: false,
  })
}
export const useResellersQuery = () =>
  useQuery({
    queryKey: ['resellers'],
    queryFn: () => read('/reseller', z.array(ResellerSchema)),
  })
export const useResellerQuery = (id: number) =>
  useQuery({
    queryKey: ['reseller', id],
    queryFn: () => read(`/reseller/${id}`, ResellerSchema),
    enabled: id > 0,
  })
export const useBillingBook = (id?: number, enabled = true) =>
  useQuery({
    queryKey: ['billing', id ?? 'me'],
    queryFn: () =>
      read(id ? `/billing/account/${id}` : '/billing/me', BillingBookSchema),
    enabled,
    refetchInterval: 15000,
  })
export const useReceiptHistory = (id?: number) =>
  useQuery({
    queryKey: ['receipts', id ?? 'me'],
    queryFn: () =>
      read(
        id ? `/billing/account/${id}/receipts` : '/billing/me/receipts',
        z.array(ReceiptSchema)
      ),
    refetchInterval: 15000,
  })
export const usePaymentInstructions = () =>
  useQuery({
    queryKey: ['payment_instructions'],
    queryFn: () => read('/billing/instructions', InstructionsSchema),
  })
export function useResellerAction<T>(
  action: (input: T) => Promise<unknown>,
  message: string
) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: action,
    onSuccess: () => {
      for (const key of [
        'resellers',
        'reseller',
        'billing',
        'receipts',
        'payment_instructions',
        'peer_purchases',
        'router_migrations',
        'my_peers',
        'peers_list',
      ])
        client.invalidateQueries({ queryKey: [key] })
      toast.success(message)
    },
  })
}
export async function downloadPrivate(path: string, name: string) {
  const { data } = await axiosInstance.get(path, { responseType: 'blob' })
  const url = URL.createObjectURL(data),
    link = document.createElement('a')
  link.href = url
  link.download = name
  link.click()
  URL.revokeObjectURL(url)
}
