import {
  LoginRequest,
  loginResponse,
  LoginResponse,
  totpConfirmResponse,
  TotpConfirm,
  TotpConfirmRequest,
  TotpDisableRequest,
  totpSetupResponse,
  TotpSetup,
  totpStatusResponse,
  TotpStatus,
  UpdateProfileRequest,
  Verify2FARequest,
} from '@/schema/authentication.ts'
import axiosInstance from '@/api/axios-instance.ts'

export const login = async (login: LoginRequest): Promise<LoginResponse> => {
  const { data } = await axiosInstance.post('/auth/login', login)
  const parsed = loginResponse.parse(data)
  return parsed.data
}

export const verify2FA = async (
  payload: Verify2FARequest
): Promise<LoginResponse> => {
  const { data } = await axiosInstance.post('/auth/totp/verify', payload)
  const parsed = loginResponse.parse(data)
  return parsed.data
}

export const updateProfile = async (
  profile: UpdateProfileRequest
): Promise<void> => {
  await axiosInstance.put('/auth/profile', profile)
}

export const getTotpStatus = async (): Promise<TotpStatus> => {
  const { data } = await axiosInstance.get('/auth/totp')
  const parsed = totpStatusResponse.parse(data)
  return parsed.data
}

export const startTotpSetup = async (): Promise<TotpSetup> => {
  const { data } = await axiosInstance.post('/auth/totp/setup')
  const parsed = totpSetupResponse.parse(data)
  return parsed.data
}

export const confirmTotpSetup = async (
  payload: TotpConfirmRequest
): Promise<TotpConfirm> => {
  const { data } = await axiosInstance.post('/auth/totp/confirm', payload)
  const parsed = totpConfirmResponse.parse(data)
  return parsed.data
}

export const disableTotp = async (payload: TotpDisableRequest): Promise<void> => {
  await axiosInstance.post('/auth/totp/disable', payload)
}
