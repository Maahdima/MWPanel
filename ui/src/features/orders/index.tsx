import { useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { IconRefresh } from '@tabler/icons-react'
import { Loader2Icon } from 'lucide-react'
import { useOrdersQuery } from '@/hooks/orders/useOrdersQuery.ts'
import { useShopEnabled } from '@/hooks/shop/useShopStatus.ts'
import { Button } from '@/components/ui/button'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { ordersColumns } from '@/features/orders/components/orders-columns.tsx'
import { DataTableSkeleton } from '@/features/shared-components/table/data-table-skeleton.tsx'
import { DataTable } from '@/features/shared-components/table/data-table.tsx'

export default function Orders() {
  const navigate = useNavigate()
  const { enabled: shopEnabled, isLoading: shopLoading } = useShopEnabled()
  const { data, isLoading, refetch, isRefetching } = useOrdersQuery(shopEnabled)

  useEffect(() => {
    if (!shopLoading && !shopEnabled) {
      navigate({ to: '/' })
    }
  }, [navigate, shopEnabled, shopLoading])

  if (shopLoading || !shopEnabled) {
    return null
  }

  return (
    <>
      <Header fixed>
        <Search />
        <div className='ml-auto flex shrink-0 items-center gap-2 sm:gap-4'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      <Main>
        <div className='mb-2 flex flex-wrap items-center justify-between gap-2 space-y-2'>
          <div className='min-w-0'>
            <h2 className='text-xl font-semibold tracking-tight'>Orders</h2>
            <p className='text-muted-foreground'>
              Telegram VPN purchases and their payment details
            </p>
          </div>
          <Button
            variant='outline'
            className='space-x-1'
            disabled={isRefetching}
            onClick={() => refetch()}
          >
            {isRefetching ? (
              <Loader2Icon className='animate-spin' />
            ) : (
              <IconRefresh />
            )}
            <span>Refresh</span>
          </Button>
        </div>

        <div className='-mx-4 flex-1 overflow-auto px-4 py-1 lg:flex-row lg:space-y-0 lg:space-x-12'>
          {isLoading ? (
            <DataTableSkeleton columns={8} rows={4} />
          ) : (
            <DataTable data={data ?? []} columns={ordersColumns} />
          )}
        </div>
      </Main>
    </>
  )
}
