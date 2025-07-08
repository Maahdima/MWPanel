import {
  LoginRequest,
  loginResponse,
  LoginResponse,
} from '@/schema/authentication.ts'
import axiosInstance from '@/api/axios-instance.ts'

export const login = async (login: LoginRequest): Promise<LoginResponse> => {
  const { data } = await axiosInstance.post('/auth/login', login)
  const parsed = loginResponse.parse(data)
  return parsed.data
}
