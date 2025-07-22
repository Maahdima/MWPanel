import { usePeersListQuery } from '@/hooks/peers/usePeersListQuery.ts'
import { useSyncPeersMutation } from '@/hooks/peers/useSyncPeersMutation.ts'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { peersColumns } from '@/features/peers/components/peers-columns.tsx'
import { DataTable } from '@/features/shared-components/table/data-table.tsx'
import { PeersDialogs } from './components/peers-dialogs.tsx'
import { PeersPrimaryButtons } from './components/peers-primary-buttons.tsx'
import PeersProvider from './context/peers-context.tsx'

export default function Peers() {
  const {
    data: peersList,
    isLoading: isPeersListLoading,
    refetch: refetchPeersList,
    isRefetching: isPeersListRefetching,
  } = usePeersListQuery()

  const { mutate: syncPeers, isPending: isPeersSyncing } =
    useSyncPeersMutation()

  return (
    <PeersProvider>
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
            <h2 className='text-2xl font-bold tracking-tight'>Peers</h2>
            <p className='text-muted-foreground'>
              Manage your wireguard configs here
            </p>
          </div>
          <PeersPrimaryButtons
            refetchPeersList={refetchPeersList}
            isPeersListRefetching={isPeersListRefetching}
            syncPeers={syncPeers}
            isPeersSyncing={isPeersSyncing}
          />
        </div>

        <div className='-mx-4 flex-1 overflow-auto px-4 py-1 lg:flex-row lg:space-y-0 lg:space-x-12'>
          {isPeersListLoading ? (
            // TODO : skeleton
            <p>Loading...</p>
          ) : (
            <DataTable data={peersList ?? []} columns={peersColumns} />
          )}
        </div>
      </Main>

      <PeersDialogs />
    </PeersProvider>
  )
}
