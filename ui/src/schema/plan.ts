import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response.ts'

export const PlanSchema = z.object({
  id: z.number(),
  name: z.string(),
  fa_name: z.string(),
  traffic_gb: z.number(),
  duration_days: z.number(),
  enabled: z.boolean(),
  sort_order: z.number(),
  price_tomans: z.number(),
})

export const UpsertPlanSchema = z.object({
  name: z.string().trim().min(1, 'Name is required'),
  fa_name: z.string().trim().min(1, 'Persian name is required'),
  traffic_gb: z.coerce.number().int().positive('Traffic must be at least 1 GB'),
  duration_days: z.coerce
    .number()
    .int()
    .positive('Duration must be at least 1 day'),
  enabled: z.boolean(),
  sort_order: z.coerce.number().int().min(0),
})

export const PlansSchema = z.array(PlanSchema).nullable()
export const PlansResponseSchema = createApiResponseSchema(PlansSchema)

export type Plan = z.infer<typeof PlanSchema>
export type UpsertPlan = z.infer<typeof UpsertPlanSchema>
