import { useServersListQuery } from '@/hooks/servers/useServersListQuery.ts'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { poolsColumns } from '@/features/pools/components/pools-columns.tsx'
import { PoolsDialogs } from '@/features/pools/components/pools-dialogs.tsx'
import { PoolsPrimaryButtons } from '@/features/pools/components/pools-primary-buttons.tsx'
import PoolsProvider from '@/features/pools/context/pools-context.tsx'
import { DataTableSkeleton } from '@/features/shared-components/table/data-table-skeleton.tsx'
import { DataTable } from '@/features/shared-components/table/data-table.tsx'

export default function Pools() {
  const {
    data: serversList,
    isLoading: isServersListLoading,
    refetch: refetchServersList,
    isRefetching: isServersListRefetching,
  } = useServersListQuery()

  return (
    <PoolsProvider>
      <Header fixed>
        <Search />
        <div className='ml-auto flex items-center space-x-4'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      <Main>
        <div className='mb-2 flex flex-wrap items-center justify-between space-y-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>IP Pools</h2>
            <p className='text-muted-foreground'>
              Manage your mikrotik ip pools here
            </p>
          </div>
          <PoolsPrimaryButtons
            refetchServersList={refetchServersList}
            isServersListRefetching={isServersListRefetching}
          />
        </div>

        <div className='-mx-4 flex-1 overflow-auto px-4 py-1 lg:flex-row lg:space-y-0 lg:space-x-12'>
          {isServersListLoading ? (
            <DataTableSkeleton columns={7} rows={1} />
          ) : (
            <DataTable data={serversList ?? []} columns={poolsColumns} />
          )}
        </div>
      </Main>

      <PoolsDialogs />
    </PoolsProvider>
  )
}
