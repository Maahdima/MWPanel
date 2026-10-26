import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response'
import {
  BillingBookSchema,
  InstructionsSchema,
  PrincipalSchema,
  ReceiptSchema,
  type ReceiptReviewInput,
  type BillingActionInput,
} from '@/schema/billing'
import axiosInstance from './axios-instance'

const accountPath = (id?: number) =>
  id ? `/billing/account/${id}` : '/billing/me'
export async function fetchPrincipal() {
  const { data } = await axiosInstance.get('/auth/me')
  return createApiResponseSchema(PrincipalSchema).parse(data).data
}
export async function fetchBillingBook(id?: number) {
  const { data } = await axiosInstance.get(accountPath(id))
  return createApiResponseSchema(BillingBookSchema).parse(data).data
}
export async function fetchReceiptHistory(id?: number) {
  const { data } = await axiosInstance.get(`${accountPath(id)}/receipts`)
  return createApiResponseSchema(z.array(ReceiptSchema)).parse(data).data
}
export async function fetchPaymentInstructions() {
  const { data } = await axiosInstance.get('/billing/instructions')
  return createApiResponseSchema(InstructionsSchema).parse(data).data
}
export async function submitReceipt(body: FormData, key: string) {
  await axiosInstance.post('/billing/me/receipts', body, {
    headers: { 'Idempotency-Key': key },
  })
}
export async function reviewReceipt({
  id,
  action,
  reason,
}: ReceiptReviewInput) {
  await axiosInstance.post(`/billing/receipts/${id}/${action}`, { reason })
}
export async function postBillingAction(
  id: number,
  { action, data, key }: BillingActionInput
) {
  await axiosInstance.post(`${accountPath(id)}/${action}`, data, {
    headers: { 'Idempotency-Key': key },
  })
}
export async function updatePaymentInstructions(data: Record<string, string>) {
  await axiosInstance.put('/billing/instructions', data)
}
