import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response.ts'

export const SyncInterfacePreviewSchema = z.object({
  id: z.string(),
  disabled: z.boolean(),
  comment: z.string().nullable(),
  name: z.string(),
  listen_port: z.string(),
  mtu: z.string(),
  is_running: z.boolean(),
})

export const SyncPeerPreviewSchema = z.object({
  id: z.string(),
  disabled: z.boolean(),
  comment: z.string().nullable(),
  name: z.string(),
  interface: z.string(),
  allowed_address: z.string(),
})

export const SyncInterfacesResponseSchema = createApiResponseSchema(
  z.array(SyncInterfacePreviewSchema)
)

export const SyncPeersResponseSchema = createApiResponseSchema(
  z.array(SyncPeerPreviewSchema)
)

export const SyncInterfacesRequestSchema = z.object({
  interface_ids: z.array(z.string()),
})

export const SyncPeersRequestSchema = z.object({
  peer_ids: z.array(z.string()),
})

export type SyncInterfacePreview = z.infer<typeof SyncInterfacePreviewSchema>
export type SyncPeerPreview = z.infer<typeof SyncPeerPreviewSchema>
export type SyncInterfacesRequest = z.infer<typeof SyncInterfacesRequestSchema>
export type SyncPeersRequest = z.infer<typeof SyncPeersRequestSchema>

