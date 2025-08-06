import {
  IconCloudDown,
  IconRefresh,
  IconRestore,
  IconUserPlus,
} from '@tabler/icons-react'
import { Loader2Icon } from 'lucide-react'
import { cn } from '@/lib/utils'
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
    <div className='flex flex-wrap items-center gap-3'>
      <div className='inline-flex w-fit -space-x-px rounded-md shadow-xs rtl:space-x-reverse'>
        <Button
          variant='outline'
          className={cn(
            'rounded-none rounded-s-md shadow-none transition-all focus-visible:z-10',
            isPeersSyncing && 'cursor-not-allowed opacity-70'
          )}
          disabled={isPeersSyncing}
          onClick={syncPeers}
        >
          {isPeersSyncing ? (
            <Loader2Icon className='h-4 w-4 animate-spin' />
          ) : (
            <IconCloudDown className='h-4 w-4' />
          )}
          <span className='text-sm font-medium'>Sync</span>
        </Button>

        <Button
          variant='outline'
          className={cn(
            'rounded-none rounded-e-md shadow-none focus-visible:z-10',
            isPeersListRefetching && 'cursor-not-allowed opacity-70'
          )}
          disabled={isPeersListRefetching}
          onClick={refetchPeersList}
        >
          {isPeersListRefetching ? (
            <Loader2Icon className='h-4 w-4 animate-spin' />
          ) : (
            <IconRefresh className='h-4 w-4' />
          )}
          <span className='text-sm font-medium'>Refresh</span>
        </Button>
      </div>

      <Button
        variant='outline'
        className='gap-2 border-amber-500 text-amber-600 transition-all hover:bg-amber-100/60 dark:border-amber-400 dark:text-amber-400 dark:hover:bg-amber-400/10'
      >
        <IconRestore className='h-4 w-4' />
        <span className='text-sm font-medium'>Reset Usages</span>
      </Button>

      <Button onClick={() => setOpen('add')} className='gap-2 transition-all'>
        <span className='text-sm font-medium'>Add Peer</span>
        <IconUserPlus className='h-4 w-4' />
      </Button>
    </div>
  )
}
