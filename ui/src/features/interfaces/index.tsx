import { useInterfacesListQuery } from '@/hooks/interfaces/useInterfacesListQuery.ts'
import { useSyncInterfacesMutation } from '@/hooks/interfaces/useSyncInterfacesMutation.ts'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { interfacesColumns } from '@/features/interfaces/components/interfaces-columns.tsx'
import { InterfacesDialogs } from '@/features/interfaces/components/interfaces-dialogs.tsx'
import { InterfacesPrimaryButtons } from '@/features/interfaces/components/interfaces-primary-buttons.tsx'
import InterfacesProvider from '@/features/interfaces/context/interfaces-context.tsx'
import { DataTableSkeleton } from '@/features/shared-components/table/data-table-skeleton.tsx'
import { DataTable } from '@/features/shared-components/table/data-table.tsx'

export default function Interfaces() {
  const {
    data: interfacesList,
    isLoading: isInterfacesListLoading,
    refetch: refetchInterfacesList,
    isRefetching: isInterfacesListRefetching,
  } = useInterfacesListQuery()

  const { mutate: syncInterfaces, isPending: isInterfacesSyncing } =
    useSyncInterfacesMutation()

  return (
    <InterfacesProvider>
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
            <h2 className='text-2xl font-bold tracking-tight'>Interfaces</h2>
            <p className='text-muted-foreground'>Manage your interfaces here</p>
          </div>
          <InterfacesPrimaryButtons
            refetchInterfacesList={refetchInterfacesList}
            isInterfacesListRefetching={isInterfacesListRefetching}
            syncInterfaces={syncInterfaces}
            isInterfacesSyncing={isInterfacesSyncing}
          />
        </div>

        <div className='-mx-4 flex-1 overflow-auto px-4 py-1 lg:flex-row lg:space-y-0 lg:space-x-12'>
          {isInterfacesListLoading ? (
            <DataTableSkeleton columns={6} rows={2} />
          ) : (
            <DataTable
              data={interfacesList ?? []}
              columns={interfacesColumns}
            />
          )}
        </div>
      </Main>

      <InterfacesDialogs />
    </InterfacesProvider>
  )
}
