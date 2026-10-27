import { useEffect, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { IconPlus, IconRefresh } from '@tabler/icons-react'
import { Loader2Icon } from 'lucide-react'
import { usePlansQuery } from '@/hooks/plans/usePlansQuery.ts'
import { useShopEnabled } from '@/hooks/shop/useShopStatus.ts'
import { Button } from '@/components/ui/button'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { BotSales } from '@/features/plans/components/bot-sales.tsx'
import { PlanFormDialog } from '@/features/plans/components/plan-form-dialog.tsx'
import { plansColumns } from '@/features/plans/components/plans-columns.tsx'
import { DataTableSkeleton } from '@/features/shared-components/table/data-table-skeleton.tsx'
import { DataTable } from '@/features/shared-components/table/data-table.tsx'

export default function Plans() {
  const navigate = useNavigate()
  const { enabled: shopEnabled, isLoading: shopLoading } = useShopEnabled()
  const { data, isLoading, refetch, isRefetching } = usePlansQuery(shopEnabled)
  const [open, setOpen] = useState(false)

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
            <h2 className='text-xl font-semibold tracking-tight'>Plans</h2>
            <p className='text-muted-foreground'>
              Shop plans shown in the panel and in the Telegram bot
            </p>
          </div>
          <div className='flex flex-wrap items-center gap-2'>
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
            <Button className='space-x-1' onClick={() => setOpen(true)}>
              <IconPlus />
              <span>Add plan</span>
            </Button>
          </div>
        </div>

        <BotSales />

        <div className='-mx-4 flex-1 overflow-auto px-4 py-1 lg:flex-row lg:space-y-0 lg:space-x-12'>
          {isLoading ? (
            <DataTableSkeleton columns={6} rows={4} />
          ) : (
            <DataTable data={data ?? []} columns={plansColumns} />
          )}
        </div>
      </Main>
      <PlanFormDialog open={open} onOpenChange={setOpen} />
    </>
  )
}
