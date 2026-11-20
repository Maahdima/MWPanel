import { z } from 'zod'
import { createApiResponseSchema } from './api-response'

export const PeerFormDefaultsSchema = z.object({
  default_expire_days: z.number().int().positive().nullable(),
})

export const PeerFormDefaultsResponseSchema = createApiResponseSchema(
  PeerFormDefaultsSchema
)

export type PeerFormDefaults = z.infer<typeof PeerFormDefaultsSchema>
