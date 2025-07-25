import { StrictMode } from 'react'
import ReactDOM from 'react-dom/client'
import { AxiosError } from 'axios'
import {
  QueryCache,
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query'
import { createRouter, RouterProvider } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/authStore'
import { handleServerError } from '@/utils/handle-server-error'
import { FontProvider } from './context/font-context'
import { ThemeProvider } from './context/theme-context'
import './index.css'
import { routeTree } from './routeTree.gen'

export const router = createRouter({
  routeTree,
  context: {} as { queryClient: QueryClient },
  defaultPreload: 'intent',
  defaultPreloadStaleTime: 0,
})

// Global HTTP error handler
const handleGlobalHttpError = (error: unknown) => {
  if (error instanceof AxiosError) {
    const status = error.response?.status ?? 0

    if (status === 401) {
      toast.error('Session expired!', { duration: 5000 })
      useAuthStore.getState().auth.reset()
      router.navigate({ to: '/sign-in' })
    }

    if (status === 403) {
      toast.error('Access denied!', { duration: 5000 })
      router.navigate({ to: '/403', replace: true })
    }

    if (status === 500) {
      toast.error('Internal Server Error!', { duration: 5000 })
      router.navigate({ to: '/500' })
    }
  }
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        if (import.meta.env.DEV)
          if (failureCount >= 0 && import.meta.env.DEV) return false
        if (failureCount > 3 && import.meta.env.PROD) return false

        return !(
          error instanceof AxiosError &&
          [401, 403].includes(error.response?.status ?? 0)
        )
      },
      refetchOnWindowFocus: import.meta.env.PROD,
      staleTime: 10 * 1000,
    },
    mutations: {
      onError: (error) => {
        handleGlobalHttpError(error)
        handleServerError(error)

        if (error instanceof AxiosError && error.response?.status === 304) {
          toast.error('Content not modified!', { duration: 5000 })
        }
      },
    },
  },
  queryCache: new QueryCache({
    onError: handleGlobalHttpError,
  }),
})

router.update({
  context: { queryClient },
})

const rootElement = document.getElementById('root')!
if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement)
  root.render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider defaultTheme='dark' storageKey='vite-ui-theme'>
          <FontProvider>
            <RouterProvider router={router} />
          </FontProvider>
        </ThemeProvider>
      </QueryClientProvider>
    </StrictMode>
  )
}
