import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response.ts'

export const IPPoolSchema = z.object({
  id: z.number(),
  name: z.string(),
  start_ip: z.string(),
  end_ip: z.string(),
  total_ip: z.number(),
  used_ip: z.number(),
  remaining_ip: z.number(),
  last_used_ip: z.string().nullable().optional(),
})

export const IPPoolsSchema = z.array(IPPoolSchema).nullable()

export const IPPoolResponseSchema = createApiResponseSchema(IPPoolSchema)
export const IPPoolsResponseSchema = createApiResponseSchema(IPPoolsSchema)

export const CreateIPPoolSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  interface_id: z.number().min(1, 'Interface ID is required'),
  start_ip: z.string().min(1, 'Start IP is required'),
  end_ip: z.string().min(1, 'End IP is required'),
})

export const UpdateIPPoolSchema = z.object({
  id: z.number().min(1, 'ID is required'),
  name: z.string().min(1, 'Name is required'),
  start_ip: z.string().min(1, 'Start IP is required'),
  end_ip: z.string().min(1, 'End IP is required'),
})

export type IPPool = z.infer<typeof IPPoolSchema>
export type CreateIPPoolRequest = z.infer<typeof CreateIPPoolSchema>
export type UpdateIPPoolRequest = z.infer<typeof UpdateIPPoolSchema>
