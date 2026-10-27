import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response.ts'

export const ShopSalesSchema = z.object({
  allow_new: z.boolean(),
  allow_renew: z.boolean(),
  allow_referral: z.boolean(),
  require_referral: z.boolean(),
})

export const ShopSalesResponseSchema = createApiResponseSchema(ShopSalesSchema)

export type ShopSales = z.infer<typeof ShopSalesSchema>
