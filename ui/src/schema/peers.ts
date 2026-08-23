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
  telegram_username: z.string().nullable(),
  telegram_linked: z.boolean().default(false),
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

export const PeerAllowedAddressSchema = z.object({
  allowed_address: z.string(),
})

export const PeerCredentialsSchema = z.object({
  private_key: z.string(),
  public_key: z.string(),
})

export const PeerShareSchema = z.object({
  is_shared: z.boolean(),
  uuid: z.string().nullable(),
  expire_time: z.string().nullable(),
})

export const PeerStatsSchema = z.object({
  name: z.string(),
  uuid: z.string(),
  expire_time: z.string().nullable(),
  traffic_limit: z.string().nullable(),
  download_usage: z.string(),
  upload_usage: z.string(),
  total_usage: z.string(),
  usage_percent: z.string().nullable(),
  is_online: z.boolean(),
})

export const PeerResponseSchema = createApiResponseSchema(PeerSchema)
export const PeersResponseSchema = createApiResponseSchema(PeersSchema)
export const PeerAllowedAddressResponseSchema = createApiResponseSchema(
  PeerAllowedAddressSchema
)
export const PeerCredentialsResponseSchema = createApiResponseSchema(
  PeerCredentialsSchema
)
export const PeerSessionStatusEnum = z.enum(['ongoing', 'ended'])

export const PeerSessionSchema = z.object({
  id: z.number(),
  status: PeerSessionStatusEnum,
  connected_at: z.string(),
  disconnected_at: z.string().nullable(),
  duration: z.number(),
  duration_label: z.string(),
  endpoint: z.string().nullable(),
  download_usage: z.string(),
  upload_usage: z.string(),
  total_usage: z.string(),
})

export const PeerSessionsSchema = z.object({
  peer_id: z.number(),
  peer_name: z.string(),
  total_sessions: z.number(),
  ongoing_sessions: z.number(),
  total_duration: z.number(),
  total_duration_label: z.string(),
  total_download_usage: z.string(),
  total_upload_usage: z.string(),
  total_usage: z.string(),
  sessions: z.array(PeerSessionSchema),
})

export const PeerShareResponseSchema = createApiResponseSchema(PeerShareSchema)
export const PeerStatsResponseSchema = createApiResponseSchema(PeerStatsSchema)
export const PeerSessionsResponseSchema =
  createApiResponseSchema(PeerSessionsSchema)

export const FetchPeerAllowedAddressSchema = z.object({
  interface_id: z.number().int().positive(),
})

export const CreatePeerSchema = z.object({
  comment: z.string().optional().nullable(),
  name: z.string().min(1, 'Name is required'),
  interface_id: z.number().min(1, 'Interface ID is required'),
  private_key: z.string().min(1, 'Private Key is required'),
  public_key: z.string().min(1, 'Public Key is required'),
  allowed_address: z.string().min(1, 'Allowed Address is required'),
  preshared_key: z.string().optional().nullable(),
  persistent_keepalive: z.string().optional().nullable(),
  endpoint: z.string().min(1, 'Endpoint is required'),
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
export type FetchPeerAllowedAddress = z.infer<
  typeof FetchPeerAllowedAddressSchema
>
export type CreatePeerRequest = z.infer<typeof CreatePeerSchema>
export type UpdatePeerRequest = z.infer<typeof UpdatePeerSchema>
export type UpdatePeerShareExpireRequest = z.infer<
  typeof UpdatePeerShareExpireSchema
>
export type PeerAllowedAddress = z.infer<typeof PeerAllowedAddressSchema>
export type PeerCredentials = z.infer<typeof PeerCredentialsSchema>
export type PeerShare = z.infer<typeof PeerShareSchema>
export type PeerStats = z.infer<typeof PeerStatsSchema>
export type PeerSession = z.infer<typeof PeerSessionSchema>
export type PeerSessions = z.infer<typeof PeerSessionsSchema>
export type PeerSessionStatus = z.infer<typeof PeerSessionStatusEnum>
