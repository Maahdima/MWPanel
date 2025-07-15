import { useServersListQuery } from '@/hooks/servers/useServersListQuery.ts'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { serversColumns } from '@/features/servers/components/servers-columns.tsx'
import { ServersDialogs } from '@/features/servers/components/servers-dialogs.tsx'
import { ServersPrimaryButtons } from '@/features/servers/components/servers-primary-buttons.tsx'
import ServersProvider from '@/features/servers/context/servers-context.tsx'
import { DataTable } from '@/features/shared-components/table/data-table.tsx'

export default function Servers() {
  const {
    data: serversList,
    isLoading: isServersListLoading,
    refetch: refetchServersList,
    isRefetching: isServersListRefetching,
  } = useServersListQuery()

  return (
    <ServersProvider>
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
            <h2 className='text-2xl font-bold tracking-tight'>Servers</h2>
            <p className='text-muted-foreground'>
              Manage your mikrotik servers here
            </p>
          </div>
          <ServersPrimaryButtons
            serversLength={serversList?.length ?? 0}
            refetchServersList={refetchServersList}
            isServersListRefetching={isServersListRefetching}
          />
        </div>

        <div className='-mx-4 flex-1 overflow-auto px-4 py-1 lg:flex-row lg:space-y-0 lg:space-x-12'>
          {isServersListLoading ? (
            // TODO : skeleton
            <p>Loading...</p>
          ) : (
            <DataTable data={serversList ?? []} columns={serversColumns} />
          )}
        </div>
      </Main>

      <ServersDialogs />
    </ServersProvider>
  )
}
