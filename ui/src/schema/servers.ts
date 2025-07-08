import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response.ts'

export const ServerStatusEnum = z.enum(['available', 'not_available'])

export const ServerSchema = z.object({
  id: z.number(),
  comment: z.string().nullable(),
  name: z.string(),
  ip_address: z.string(),
  api_port: z.string(),
  status: ServerStatusEnum,
  is_active: z.boolean(),
})

export const CreateServerSchema = z.object({
  comment: z.string().optional().nullable(),
  name: z.string().min(1, 'Name is required'),
  ip_address: z.string().min(1, 'IP Address is required'),
  api_port: z.string().min(1, 'Rest API Port is required'),
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(1, 'Password is required'),
})

export const UpdateServerSchema = z.object({
  id: z.number().int().positive(),
  comment: z.string().optional().nullable(),
  name: z.string().optional().nullable(),
  ip_address: z.string().optional().nullable(),
  api_port: z.string().optional().nullable(),
  username: z.string().optional().nullable(),
  password: z.string().optional().nullable(),
})

export const ServersSchema = z.array(ServerSchema).nullable()

export const ServerResponseSchema = createApiResponseSchema(ServerSchema)
export const ServersResponseSchema = createApiResponseSchema(ServersSchema)

export type Server = z.infer<typeof ServerSchema>
export type ServerStatus = z.infer<typeof ServerStatusEnum>
export type CreateServerRequest = z.infer<typeof CreateServerSchema>
export type UpdateServerRequest = Partial<Server> & { id: number }
export type ServerResponse = z.infer<typeof ServerResponseSchema>
export type ServersResponse = z.infer<typeof ServersResponseSchema>
