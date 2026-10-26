import Cookies from 'js-cookie'
import {
  Navigate,
  Outlet,
  useLocation,
  useRouter,
} from '@tanstack/react-router'
import { useAuthStore } from '@/stores/authStore'
import { cn } from '@/lib/utils'
import { SearchProvider } from '@/context/search-context'
import {
  usePrincipal,
  useBillingBook,
} from '@/hooks/resellers/useResellerQueries'
import { Button } from '@/components/ui/button'
import { SidebarProvider } from '@/components/ui/sidebar'
import { AppSidebar } from '@/components/layout/app-sidebar'
import SkipToMain from '@/components/skip-to-main'

interface Props {
  children?: React.ReactNode
}

export function AuthenticatedLayout({ children }: Props) {
  const defaultOpen = Cookies.get('sidebar_state') !== 'false'
  const token = useAuthStore((state) => state.auth.accessToken)
  const principal = usePrincipal()
  const location = useLocation()
  const basepath = useRouter().options.basepath?.replace(/\/$/, '') ?? ''
  const routePath =
    basepath && location.pathname.startsWith(basepath + '/')
      ? location.pathname.slice(basepath.length)
      : location.pathname
  const reseller = principal.data?.role === 'reseller'
  const book = useBillingBook(undefined, reseller)
  if (!token) return <Navigate to='/sign-in' replace />
  if (principal.isPending || (reseller && book.isPending))
    return <div className='p-8'>Loading account…</div>
  if (principal.isError || (reseller && book.isError))
    return (
      <div className='p-8'>
        Unable to load account.{' '}
        <Button
          onClick={() => {
            principal.refetch()
            book.refetch()
          }}
        >
          Retry
        </Button>
      </div>
    )
  if (
    reseller &&
    !['/billing', '/peers'].includes(routePath.replace(/\/$/, ''))
  )
    return <Navigate to='/billing' replace />
  return (
    <SearchProvider>
      <SidebarProvider defaultOpen={defaultOpen}>
        <SkipToMain />
        <AppSidebar />
        <div
          id='content'
          className={cn(
            'ml-auto w-full max-w-full min-w-0',
            'md:peer-data-[state=collapsed]:w-[calc(100%-var(--sidebar-width-icon)-1rem)]',
            'md:peer-data-[state=expanded]:w-[calc(100%-var(--sidebar-width))]',
            'sm:transition-[width] sm:duration-200 sm:ease-linear',
            'flex h-svh flex-col',
            'group-data-[scroll-locked=1]/body:h-full',
            'has-[main.fixed-main]:group-data-[scroll-locked=1]/body:h-svh'
          )}
        >
          {children ? children : <Outlet />}
        </div>
      </SidebarProvider>
    </SearchProvider>
  )
}
