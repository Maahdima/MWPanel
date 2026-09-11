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
  requires_2fa: z.boolean().optional().default(false),
  temp_token: z.string().optional(),
  user_id: z.number().optional(),
  username: z.string().optional(),
  access_token: z.string().optional(),
  refresh_token: z.string().optional(),
  expires_in: z.number().optional(),
})

export const loginResponse = createApiResponseSchema(loginResponseSchema)

export const verify2FARequestSchema = z.object({
  temp_token: z.string().min(1),
  code: z.string().min(1, { message: 'Please enter your authentication code' }),
})

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

export const totpStatusSchema = z.object({
  enabled: z.boolean(),
})

export const totpStatusResponse = createApiResponseSchema(totpStatusSchema)

export const totpSetupSchema = z.object({
  secret: z.string(),
  otpauth_url: z.string(),
  qr_code_data_url: z.string(),
})

export const totpSetupResponse = createApiResponseSchema(totpSetupSchema)

export const totpConfirmRequestSchema = z.object({
  code: z.string().min(6, { message: 'Enter the 6-digit code from your app' }),
})

export const totpConfirmSchema = z.object({
  recovery_codes: z.array(z.string()),
})

export const totpConfirmResponse = createApiResponseSchema(totpConfirmSchema)

export const totpDisableRequestSchema = z.object({
  password: z.string().min(1, { message: 'Please enter your password' }),
  code: z.string().min(1, { message: 'Please enter your authentication code' }),
})

export type LoginRequest = z.infer<typeof loginRequestSchema>
export type LoginResponse = z.infer<typeof loginResponseSchema>
export type Verify2FARequest = z.infer<typeof verify2FARequestSchema>
export type UpdateProfileRequest = z.infer<typeof updateProfileSchema>
export type TotpStatus = z.infer<typeof totpStatusSchema>
export type TotpSetup = z.infer<typeof totpSetupSchema>
export type TotpConfirmRequest = z.infer<typeof totpConfirmRequestSchema>
export type TotpConfirm = z.infer<typeof totpConfirmSchema>
export type TotpDisableRequest = z.infer<typeof totpDisableRequestSchema>
