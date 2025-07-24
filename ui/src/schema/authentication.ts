import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response.ts'

export const loginRequestSchema = z.object({
  username: z.string().min(1, { message: 'Please enter your username' }),
  password: z
    .string()
    .min(1, {
      message: 'Please enter your password',
    })
    .min(7, {
      message: 'Password must be at least 7 characters long',
    }),
})

export const loginResponseSchema = z.object({
  user_id: z.number(),
  username: z.string(),
  access_token: z.string(),
  refresh_token: z.string(),
  expires_in: z.number(),
})

export const loginResponse = createApiResponseSchema(loginResponseSchema)

export const updateProfileSchema = z.object({
  old_username: z.string().min(1, { message: 'Please enter your username' }),
  old_password: z.string().min(1, {
    message: 'Please enter your password',
  }),
  new_username: z.string().optional(),
  new_password: z
    .string()
    .min(8, { message: 'Password must be at least 8 characters long' })
    .optional()
    .or(z.literal('')),
})

export type LoginRequest = z.infer<typeof loginRequestSchema>
export type LoginResponse = z.infer<typeof loginResponseSchema>
export type UpdateProfileRequest = z.infer<typeof updateProfileSchema>
