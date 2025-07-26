import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response.ts'

export const PeerStatusEnum = z.enum([
  'active',
  'inactive',
  'expired',
  'suspended',
])

export const PeerSchema = z.object({
  id: z.number(),
  uuid: z.string(),
  disabled: z.boolean(),
  comment: z.string().nullable(),
  name: z.string(),
  interface: z.string(),
  allowed_address: z.string(),
  traffic_limit: z.string().nullable(),
  expire_time: z.string().nullable(),
  download_bandwidth: z.string().nullable(),
  upload_bandwidth: z.string().nullable(),
  total_usage: z.string(),
  status: z.array(PeerStatusEnum),
  is_online: z.boolean(),
  is_shared: z.boolean(),
})

export const PeersSchema = z.array(PeerSchema).nullable()

export const PeerKeysSchema = z.object({
  private_key: z.string(),
  public_key: z.string(),
})

export const PeerShareSchema = z.object({
  is_shared: z.boolean(),
  share_link: z.string().nullable(),
  expire_time: z.string().nullable(),
})

export const PeerStatsSchema = z.object({
  name: z.string(),
  expire_time: z.string().nullable(),
  traffic_limit: z.string().nullable(),
  download_usage: z.string(),
  upload_usage: z.string(),
  total_usage: z.string(),
  usage_percent: z.string().nullable(),
})

export const PeerResponseSchema = createApiResponseSchema(PeerSchema)
export const PeersResponseSchema = createApiResponseSchema(PeersSchema)
export const PeerKeysResponseSchema = createApiResponseSchema(PeerKeysSchema)
export const PeerShareResponseSchema = createApiResponseSchema(PeerShareSchema)
export const PeerStatsResponseSchema = createApiResponseSchema(PeerStatsSchema)

export const CreatePeerSchema = z.object({
  comment: z.string().optional().nullable(),
  name: z.string().min(1, 'Name is required'),
  interface_id: z.string().min(1, 'Interface ID is required'),
  interface_name: z.string().min(1, 'Interface Name is required'),
  private_key: z.string().min(1, 'Private Key is required'),
  public_key: z.string().min(1, 'Public Key is required'),
  allowed_address: z.string().min(1, 'Allowed Address is required'),
  preshared_key: z.string().optional().nullable(),
  persistent_keepalive: z.string().optional().nullable(),
  endpoint: z.string().min(1, 'Endpoint is required'),
  listen_port: z.string().min(1, 'Listen Port is required'),
  expire_time: z.string().optional().nullable(),
  traffic_limit: z.string().optional().nullable(),
  download_bandwidth: z.string().optional().nullable(),
  upload_bandwidth: z.string().optional().nullable(),
})

export const UpdatePeerSchema = z.object({
  id: z.number().int().positive(),
  disabled: z.boolean().optional(),
  comment: z.string().optional().nullable(),
  name: z.string().min(1, 'Name is required'),
  allowed_address: z.string().min(1, 'Allowed Address is required'),
  persistent_keepalive: z.string().optional().nullable(),
  expire_time: z.string().optional().nullable(),
  traffic_limit: z.string().optional().nullable(),
  download_bandwidth: z.string().optional().nullable(),
  upload_bandwidth: z.string().optional().nullable(),
})

export const UpdatePeerShareExpireSchema = z.object({
  id: z.number().int().positive(),
  expire_time: z.string().optional().nullable(),
})

export type Peer = z.infer<typeof PeerSchema>
export type PeerStatus = z.infer<typeof PeerStatusEnum>
export type CreatePeerRequest = z.infer<typeof CreatePeerSchema>
export type UpdatePeerRequest = z.infer<typeof UpdatePeerSchema>
export type UpdatePeerShareExpireRequest = z.infer<
  typeof UpdatePeerShareExpireSchema
>
export type PeerResponse = z.infer<typeof PeerResponseSchema>
export type PeersResponse = z.infer<typeof PeersResponseSchema>
export type PeerKeys = z.infer<typeof PeerKeysSchema>
export type PeerShare = z.infer<typeof PeerShareSchema>
export type PeerStats = z.infer<typeof PeerStatsSchema>
