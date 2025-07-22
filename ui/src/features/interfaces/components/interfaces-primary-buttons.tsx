import { IconCloudDown, IconRefresh, IconUserPlus } from '@tabler/icons-react'
import { Loader2Icon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useInterfaces } from '@/features/interfaces/context/interfaces-context.tsx'

interface Props {
  refetchInterfacesList: () => void
  isInterfacesListRefetching: boolean
  syncInterfaces: () => void
  isInterfacesSyncing: boolean
}

export function InterfacesPrimaryButtons({
  refetchInterfacesList,
  isInterfacesListRefetching,
  syncInterfaces,
  isInterfacesSyncing,
}: Props) {
  const { setOpen } = useInterfaces()

  return (
    <div className='flex gap-2'>
      <Button
        variant='outline'
        className='border-primary space-x-1 border-dashed shadow-none'
        disabled={isInterfacesSyncing}
        onClick={syncInterfaces}
      >
        <span>{isInterfacesSyncing ? 'Syncing...' : 'Sync Interfaces'}</span>{' '}
        {isInterfacesSyncing ? (
          <Loader2Icon className='animate-spin' />
        ) : (
          <IconCloudDown />
        )}
      </Button>
      <Button
        variant='outline'
        className='space-x-1'
        disabled={isInterfacesListRefetching}
        onClick={refetchInterfacesList}
      >
        <span>{isInterfacesListRefetching ? 'Please wait' : 'Refresh'}</span>{' '}
        {isInterfacesListRefetching ? (
          <Loader2Icon className='animate-spin' />
        ) : (
          <IconRefresh />
        )}
      </Button>
      <Button className='space-x-1' onClick={() => setOpen('add')}>
        <span>Add New Interface</span> <IconUserPlus size={18} />
      </Button>
    </div>
  )
}
