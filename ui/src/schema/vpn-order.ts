import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response.ts'

export const VpnOrderSchema = z.object({
  id: z.number(),
  tracking_code: z.string(),
  chat_id: z.string(),
  username: z.string().nullable().optional(),
  kind: z.string(),
  config_name: z.string(),
  plan_title: z.string(),
  plan_name: z.string().optional(),
  traffic_gb: z.number(),
  duration_days: z.number(),
  amount_tomans: z.number(),
  payer_card: z.string(),
  status: z.string(),
  expires_at: z.number(),
  paid_at: z.number().nullable().optional(),
  created_at: z.number(),
  failure_reason: z.string().nullable().optional(),
  manual: z.boolean().optional(),
  config_id: z.number().nullable().optional(),
  peer_name: z.string().nullable().optional(),
  peer_expire: z.string().nullable().optional(),
})

export const VpnOrdersSchema = z.array(VpnOrderSchema).nullable()
export const VpnOrdersResponseSchema = createApiResponseSchema(VpnOrdersSchema)

export type VpnOrder = z.infer<typeof VpnOrderSchema>
