import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response.ts'

export const TelegramStatusSchema = z.object({
  enabled: z.boolean(),
  bot_username: z.string().optional(),
  bot_url: z.string().optional(),
})

export const TelegramStatusResponseSchema =
  createApiResponseSchema(TelegramStatusSchema)

export type TelegramStatus = z.infer<typeof TelegramStatusSchema>

export const PeerTelegramStatusSchema = z.object({
  linked: z.boolean(),
  username: z.string().nullable().optional(),
  notify_enabled: z.boolean(),
})

export const PeerTelegramStatusResponseSchema = createApiResponseSchema(
  PeerTelegramStatusSchema
)

export type PeerTelegramStatus = z.infer<typeof PeerTelegramStatusSchema>
