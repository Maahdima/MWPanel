import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response.ts'

export const ShopStatusSchema = z.object({
  enabled: z.boolean(),
})

export const ShopStatusResponseSchema = createApiResponseSchema(ShopStatusSchema)

export type ShopStatus = z.infer<typeof ShopStatusSchema>
