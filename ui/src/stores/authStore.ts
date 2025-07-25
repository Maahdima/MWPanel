import Cookies from 'js-cookie'
import { create } from 'zustand'

interface AuthAdmin {
  user_id: number
  username: string
}

interface AuthState {
  auth: {
    admin: AuthAdmin | null
    setAdmin: (admin: AuthAdmin) => void
    accessToken: string
    setAccessToken: (accessToken: string) => void
    resetAccessToken: () => void
    reset: () => void
  }
}

export const useAuthStore = create<AuthState>()((set) => {
  const cookieState = Cookies.get('access_token')
  const initToken = cookieState ? cookieState : ''
  return {
    auth: {
      admin: null,
      setAdmin: (admin) =>
        set((state) => {
          Cookies.set('username', admin?.username)
          return { ...state, auth: { ...state.auth, admin } }
        }),
      accessToken: initToken,
      setAccessToken: (accessToken) =>
        set((state) => {
          Cookies.set('access_token', accessToken)
          return { ...state, auth: { ...state.auth, accessToken } }
        }),
      resetAccessToken: () =>
        set((state) => {
          Cookies.remove('access_token')
          return { ...state, auth: { ...state.auth, accessToken: '' } }
        }),
      reset: () =>
        set((state) => {
          Cookies.remove('access_token')
          Cookies.remove('username')
          return {
            ...state,
            auth: { ...state.auth, admin: null, accessToken: '' },
          }
        }),
    },
  }
})
