import { IconCloudDown, IconRefresh, IconUserPlus } from '@tabler/icons-react'
import { Loader2Icon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { usePeers } from '@/features/peers/context/peers-context.tsx'

interface Props {
  refetchPeersList: () => void
  isPeersListRefetching: boolean
  syncPeers: () => void
  isPeersSyncing: boolean
}

export function PeersPrimaryButtons({
  refetchPeersList,
  isPeersListRefetching,
  syncPeers,
  isPeersSyncing,
}: Props) {
  const { setOpen } = usePeers()

  return (
    <div className='flex gap-2'>
      <Button
        variant='outline'
        className='border-primary space-x-1 border-dashed shadow-none'
        disabled={isPeersSyncing}
        onClick={syncPeers}
      >
        <span>{isPeersSyncing ? 'Syncing...' : 'Sync Peers'}</span>{' '}
        {isPeersSyncing ? (
          <Loader2Icon className='animate-spin' />
        ) : (
          <IconCloudDown />
        )}
      </Button>
      <Button
        variant='outline'
        className='space-x-1'
        disabled={isPeersListRefetching}
        onClick={refetchPeersList}
      >
        <span>{isPeersListRefetching ? 'Please wait' : 'Refresh'}</span>{' '}
        {isPeersListRefetching ? (
          <Loader2Icon className='animate-spin' />
        ) : (
          <IconRefresh />
        )}
      </Button>
      {/*<Button variant='outline' className='space-x-1' disabled={true}>*/}
      {/*  <span>Export</span> <IconUpload size={18} />*/}
      {/*</Button>*/}
      {/*<Button variant='outline' className='space-x-1' disabled={true}>*/}
      {/*  <span>Import</span> <IconDownload size={18} />*/}
      {/*</Button>*/}
      <Button className='space-x-1' onClick={() => setOpen('add')}>
        <span>Add New Peer</span> <IconUserPlus size={18} />
      </Button>
    </div>
  )
}
