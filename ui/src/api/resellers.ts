import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response'
import {
  ResellerSchema,
  ResellerResourceSchema,
  PeerPurchasesSchema,
  RouterMigrationsSchema,
  type ResellerInput,
  type ResellerPeerInput,
  type PeerPurchaseInput,
  type PeerMutationInput,
  type RouterMigrationInput,
  type OwnershipInput,
} from '@/schema/billing'
import { PeersResponseSchema } from '@/schema/peers'
import axiosInstance from './axios-instance'

const resellerPath = (id?: number) => `/reseller/${id ?? 'me'}`
export async function fetchResellers() {
  const { data } = await axiosInstance.get('/reseller')
  return createApiResponseSchema(z.array(ResellerSchema)).parse(data).data
}
export async function fetchReseller(id: number) {
  const { data } = await axiosInstance.get(resellerPath(id))
  return createApiResponseSchema(ResellerSchema).parse(data).data
}
export async function fetchResellerPeers(id?: number) {
  const { data } = await axiosInstance.get(`${resellerPath(id)}/peers`)
  return PeersResponseSchema.parse(data).data ?? []
}
export async function fetchMyResources() {
  const { data } = await axiosInstance.get('/reseller/me/resources')
  return createApiResponseSchema(z.array(ResellerResourceSchema)).parse(data)
    .data
}
export async function fetchPeerPurchases(peerID: number, id?: number) {
  const { data } = await axiosInstance.get(
    `${resellerPath(id)}/peers/${peerID}/purchases`
  )
  return createApiResponseSchema(PeerPurchasesSchema).parse(data).data
}
export async function fetchRouterMigrations(id: number) {
  const { data } = await axiosInstance.get(`${resellerPath(id)}/migrations`)
  return createApiResponseSchema(RouterMigrationsSchema).parse(data).data
}
export async function saveReseller(input: ResellerInput, id?: number) {
  const { data } = id
    ? await axiosInstance.put(resellerPath(id), input)
    : await axiosInstance.post('/reseller', input)
  return createApiResponseSchema(ResellerSchema).parse(data).data
}
export async function setResellerSuspension(
  id: number,
  input: { suspended: boolean; reason: string }
) {
  await axiosInstance.post(`${resellerPath(id)}/suspension`, input)
}
export async function transferPeerOwnership(
  peerID: number,
  input: OwnershipInput
) {
  await axiosInstance.put(`/peer/${peerID}/owner`, input)
}
export async function createResellerPeer(
  input: ResellerPeerInput,
  key: string,
  id?: number
) {
  await axiosInstance.post(`${resellerPath(id)}/peers`, input, {
    headers: { 'Idempotency-Key': key },
  })
}
export async function purchasePeerOperation(
  peerID: number,
  input: PeerPurchaseInput,
  key: string,
  id?: number
) {
  await axiosInstance.post(
    `${resellerPath(id)}/peers/${peerID}/purchases`,
    input,
    { headers: { 'Idempotency-Key': key } }
  )
}
export async function mutateResellerPeer(
  peerID: number,
  input: PeerMutationInput,
  key: string,
  id?: number
) {
  await axiosInstance.post(
    `${resellerPath(id)}/peers/${peerID}/mutations`,
    input,
    { headers: { 'Idempotency-Key': key } }
  )
}
export async function beginRouterMigration(
  id: number,
  input: RouterMigrationInput,
  key: string
) {
  await axiosInstance.post(`${resellerPath(id)}/migrations`, input, {
    headers: { 'Idempotency-Key': key },
  })
}
export async function cancelRouterMigration(id: number, migrationID: string) {
  await axiosInstance.post(
    `${resellerPath(id)}/migrations/${migrationID}/cancel`,
    { reason: 'Administrator requested cancellation before routing switch' }
  )
}
